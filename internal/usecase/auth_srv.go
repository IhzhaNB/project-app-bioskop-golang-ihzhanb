package usecase

import (
	"context"
	"fmt"
	"time"

	"cinema-booking/internal/data/entity"
	"cinema-booking/internal/data/repository"
	"cinema-booking/internal/dto/request"
	"cinema-booking/internal/dto/response"
	"cinema-booking/pkg/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService interface {
	Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error)
	Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error)
	Logout(ctx context.Context, token string) error
	SendOTP(ctx context.Context, email, otpType string) error
	VerifyEmail(ctx context.Context, req *request.VerifyEmailRequest) error
}

type authService struct {
	repo   *repository.Repository
	config *utils.Config
	log    *zap.Logger
}

func NewAuthService(
	repo *repository.Repository,
	config *utils.Config,
	log *zap.Logger,
) AuthService {
	return &authService{
		repo:   repo,
		config: config,
		log:    log,
	}
}

func (s *authService) Register(ctx context.Context, req *request.RegisterRequest) (*response.AuthResponse, error) {
	// Validate input
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Register validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Check if email already exists (prevent duplicate registration)
	existingUser, err := s.repo.User.FindByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("Failed to check email", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("check email %s: %w", req.Email, err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email %s already registered", req.Email)
	}

	// Check if username already taken
	existingUser, err = s.repo.User.FindByUsername(ctx, req.Username)
	if err != nil {
		s.log.Error("Failed to check username", zap.Error(err), zap.String("username", req.Username))
		return nil, fmt.Errorf("check username %s: %w", req.Username, err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username %s already taken", req.Username)
	}

	// Hash password using bcrypt before storing
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Create user entity with UUID and timestamps
	now := time.Now()
	user := &entity.User{
		Base: entity.Base{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Username:      req.Username,
		Email:         req.Email,
		PasswordHash:  hashedPassword,
		Phone:         req.Phone,
		Role:          entity.RoleCustomer, // Default role: customer
		EmailVerified: false,               // Email not verified yet
		IsActive:      true,                // Account is active by default
	}

	// Save to database
	if err := s.repo.User.Create(ctx, user); err != nil {
		s.log.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("create user account: %w", err)
	}

	// Send verification OTP asynchronously (using goroutine)
	go s.sendVerificationOTP(user.Email) // Non-blocking

	// Create session for auto-login after registration
	session, err := s.createSession(ctx, user.ID)
	if err != nil {
		s.log.Warn("Failed to create session after register",
			zap.Error(err), zap.String("user_id", user.ID.String()))
	}

	s.log.Info("User registered",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.String("username", user.Username))

	// Convert to response DTO
	authResp := response.AuthToResponse(user, session)
	return &authResp, nil
}

func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	// Validate input
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Login validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Try to find user by email first, then by username
	var user *entity.User
	var err error

	user, err = s.repo.User.FindByEmail(ctx, req.Username)
	if err != nil {
		s.log.Error("Failed to find user by email", zap.Error(err), zap.String("identifier", req.Username))
		return nil, fmt.Errorf("find user by email %s: %w", req.Username, err)
	}

	// If not found by email, try username
	if user == nil {
		user, err = s.repo.User.FindByUsername(ctx, req.Username)
		if err != nil {
			s.log.Error("Failed to find user by username", zap.Error(err), zap.String("identifier", req.Username))
			return nil, fmt.Errorf("find user by username %s: %w", req.Username, err)
		}
	}

	// User not found
	if user == nil {
		s.log.Warn("User not found for login", zap.String("identifier", req.Username))
		return nil, fmt.Errorf("user %s not found", req.Username)
	}

	// Verify password using bcrypt compare
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		s.log.Warn("Invalid password", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("invalid password for user %s", req.Username)
	}

	// Check if account is active (not banned/deactivated)
	if !user.IsActive {
		s.log.Warn("Inactive user tried to login", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("account %s is deactivated", req.Username)
	}

	// Create new session
	session, err := s.createSession(ctx, user.ID)
	if err != nil {
		s.log.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("create session for user %s: %w", user.ID.String(), err)
	}

	s.log.Info("User logged in",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username))

	// Return response with session token
	authResp := response.AuthToResponse(user, session)
	return &authResp, nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// Parse string token to UUID
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		s.log.Warn("Invalid token format", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("invalid token format %s: %w", token, err)
	}

	// Revoke session
	if err := s.repo.Session.Revoke(ctx, tokenUUID.String()); err != nil {
		s.log.Error("Failed to revoke session", zap.Error(err), zap.String("token", token))
		return fmt.Errorf("revoke session token %s: %w", token, err)
	}

	s.log.Info("User logged out", zap.String("token", token))
	return nil
}

func (s *authService) SendOTP(ctx context.Context, email, otpType string) error {
	// Find user
	user, err := s.repo.User.FindByEmail(ctx, email)
	if err != nil {
		s.log.Error("Failed to find user for OTP", zap.Error(err), zap.String("email", email))
		return fmt.Errorf("find user for OTP %s: %w", email, err)
	}
	if user == nil {
		return fmt.Errorf("user with email %s not found", email)
	}

	// Check if already verified (for email verification)
	if otpType == string(entity.OTPTypeEmailVerification) && user.EmailVerified {
		return fmt.Errorf("email %s already verified", email)
	}

	// Generate OTP
	otpCode := utils.GenerateOTP(s.config.OTP.Length)
	expiresAt := time.Now().Add(time.Duration(s.config.OTP.ExpiryMinutes) * time.Minute)

	// Create OTP entity
	otp := &entity.OTP{
		BaseSimple: entity.BaseSimple{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		UserID:    user.ID,
		Email:     email,
		OTPCode:   otpCode,
		OTPType:   entity.OTPType(otpType),
		ExpiresAt: expiresAt,
		IsUsed:    false,
	}

	// Save OTP
	if err := s.repo.OTP.Create(ctx, otp); err != nil {
		s.log.Error("Failed to save OTP", zap.Error(err), zap.String("email", email))
		return fmt.Errorf("save OTP for %s: %w", email, err)
	}

	// Log OTP (in development)
	s.log.Info("OTP generated",
		zap.String("email", email),
		zap.String("otp_type", otpType),
		zap.Time("expires_at", expiresAt),
	)

	// Print to console for development
	fmt.Printf("\n📧 OTP for %s (%s): %s (Expires: %s)\n\n",
		email, otpType, otpCode, expiresAt.Format("15:04:05"))

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *request.VerifyEmailRequest) error {
	// Validate input
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Verify email validation failed", zap.Any("errors", errs))
		return fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Find valid OTP
	otp, err := s.repo.OTP.FindValidOTP(ctx, req.Email, req.OTP, string(entity.OTPTypeEmailVerification))
	if err != nil {
		s.log.Error("Failed to find OTP", zap.Error(err), zap.String("email", req.Email))
		return fmt.Errorf("find OTP for %s: %w", req.Email, err)
	}
	if otp == nil {
		return fmt.Errorf("invalid or expired OTP for email %s", req.Email)
	}

	// Mark OTP as used
	if err := s.repo.OTP.MarkAsUsed(ctx, otp.ID); err != nil {
		s.log.Warn("Failed to mark OTP as used", zap.Error(err), zap.String("otp_id", otp.ID.String()))
		// Continue anyway
	}

	// Find user
	user, err := s.repo.User.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		s.log.Error("User not found for verification", zap.Error(err), zap.String("email", req.Email))
		return fmt.Errorf("find user for verification %s: %w", req.Email, err)
	}

	// Update user verification status
	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	if err := s.repo.User.Update(ctx, user); err != nil {
		s.log.Error("Failed to update user verification", zap.Error(err), zap.String("user_id", user.ID.String()))
		return fmt.Errorf("update user verification %s: %w", user.ID.String(), err)
	}

	s.log.Info("Email verified",
		zap.String("email", req.Email),
		zap.String("user_id", user.ID.String()))

	return nil
}

// ==================== HELPER METHODS ====================

func (s *authService) createSession(ctx context.Context, userID uuid.UUID) (*entity.Session, error) {
	session := &entity.Session{
		BaseSimple: entity.BaseSimple{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		},
		UserID:    userID,
		Token:     uuid.New(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	if err := s.repo.Session.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return session, nil
}

func (s *authService) sendVerificationOTP(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.SendOTP(ctx, email, string(entity.OTPTypeEmailVerification)); err != nil {
		s.log.Error("Failed to send verification OTP", zap.Error(err), zap.String("email", email))
	}
}
