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
	repo   *repository.Repository // grouping userRepo, sessionRepo, & otpRepo
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
	// 1. Validasi input
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Register validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// 2. Cek email sudah terdaftar
	existingUser, err := s.repo.User.FindByEmail(ctx, req.Email)
	if err != nil {
		s.log.Error("Failed to check email", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to check email")
	}
	if existingUser != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// 3. Cek username sudah dipakai
	existingUser, err = s.repo.User.FindByUsername(ctx, req.Username)
	if err != nil {
		s.log.Error("Failed to check username", zap.Error(err), zap.String("username", req.Username))
		return nil, fmt.Errorf("failed to check username")
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already taken")
	}

	// 4. Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		s.log.Error("Failed to hash password", zap.Error(err))
		return nil, fmt.Errorf("failed to process password")
	}

	// 5. Create user entity
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
		Role:          entity.RoleCustomer,
		EmailVerified: false,
		IsActive:      true,
	}

	// 6. Save user
	if err := s.repo.User.Create(ctx, user); err != nil {
		s.log.Error("Failed to create user", zap.Error(err), zap.String("email", req.Email))
		return nil, fmt.Errorf("failed to create account")
	}

	// 7. Send OTP email (async)
	go s.sendVerificationOTP(user.Email)

	// 8. Auto login setelah register
	session, err := s.createSession(ctx, user.ID)
	if err != nil {
		s.log.Warn("Failed to create session after register",
			zap.Error(err), zap.String("user_id", user.ID.String()))
		// Continue tanpa session
	}

	s.log.Info("User registered",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email))

	return s.convertAuthResponse(user, session), nil
}

func (s *authService) Login(ctx context.Context, req *request.LoginRequest) (*response.AuthResponse, error) {
	// 1. Validasi
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Login validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// 2. Find user by username or email
	var user *entity.User
	var err error

	// Coba cari by email
	user, err = s.repo.User.FindByEmail(ctx, req.Username)
	if err != nil {
		s.log.Error("Failed to find user by email", zap.Error(err), zap.String("identifier", req.Username))
		return nil, fmt.Errorf("failed to find user")
	}

	// Jika tidak ditemukan, coba by username
	if user == nil {
		user, err = s.repo.User.FindByUsername(ctx, req.Username)
		if err != nil {
			s.log.Error("Failed to find user by username", zap.Error(err), zap.String("identifier", req.Username))
			return nil, fmt.Errorf("failed to find user")
		}
	}

	// 3. User not found
	if user == nil {
		s.log.Warn("User not found for login", zap.String("identifier", req.Username))
		return nil, fmt.Errorf("invalid credentials")
	}

	// 4. Check password
	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		s.log.Warn("Invalid password", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("invalid credentials")
	}

	// 5. Check if user is active
	if !user.IsActive {
		s.log.Warn("Inactive user tried to login", zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("account is deactivated")
	}

	// 6. Create session
	session, err := s.createSession(ctx, user.ID)
	if err != nil {
		s.log.Error("Failed to create session", zap.Error(err), zap.String("user_id", user.ID.String()))
		return nil, fmt.Errorf("failed to create session")
	}

	s.log.Info("User logged in",
		zap.String("user_id", user.ID.String()),
		zap.String("username", user.Username))

	return s.convertAuthResponse(user, session), nil
}

func (s *authService) Logout(ctx context.Context, token string) error {
	// 1. Parse token
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		s.log.Warn("Invalid token format", zap.String("token", token), zap.Error(err))
		return fmt.Errorf("invalid token format")
	}

	// 2. Revoke session
	if err := s.repo.Session.Revoke(ctx, tokenUUID.String()); err != nil {
		s.log.Error("Failed to revoke session", zap.Error(err), zap.String("token", token))
		return fmt.Errorf("failed to logout")
	}

	s.log.Info("User logged out", zap.String("token", token))
	return nil
}

func (s *authService) SendOTP(ctx context.Context, email, otpType string) error {
	// 1. Find user
	user, err := s.repo.User.FindByEmail(ctx, email)
	if err != nil {
		s.log.Error("Failed to find user for OTP", zap.Error(err), zap.String("email", email))
		return fmt.Errorf("failed to find user")
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// 2. Check if already verified (for email verification)
	if otpType == string(entity.OTPTypeEmailVerification) && user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// 3. Generate OTP
	otpCode := utils.GenerateOTP(s.config.OTP.Length)
	expiresAt := time.Now().Add(time.Duration(s.config.OTP.ExpiryMinutes) * time.Minute)

	// 4. Create OTP entity
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

	// 5. Save OTP
	if err := s.repo.OTP.Create(ctx, otp); err != nil {
		s.log.Error("Failed to save OTP", zap.Error(err), zap.String("email", email))
		return fmt.Errorf("failed to generate OTP")
	}

	// 6. Log OTP (in development)
	s.log.Info("OTP generated",
		zap.String("email", email),
		zap.String("otp_code", otpCode),
		zap.String("otp_type", otpType),
		zap.Time("expires_at", expiresAt),
	)

	// Print to console for development
	fmt.Printf("\n📧 OTP for %s (%s): %s (Expires: %s)\n\n",
		email, otpType, otpCode, expiresAt.Format("15:04:05"))

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, req *request.VerifyEmailRequest) error {
	// 1. Validasi
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Verify email validation failed", zap.Any("errors", errs))
		return fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// 2. Find valid OTP
	otp, err := s.repo.OTP.FindValidOTP(ctx, req.Email, req.OTP, string(entity.OTPTypeEmailVerification))
	if err != nil {
		s.log.Error("Failed to find OTP", zap.Error(err), zap.String("email", req.Email))
		return fmt.Errorf("failed to verify OTP")
	}
	if otp == nil {
		return fmt.Errorf("invalid or expired OTP")
	}

	// 3. Mark OTP as used
	if err := s.repo.OTP.MarkAsUsed(ctx, otp.ID); err != nil {
		s.log.Warn("Failed to mark OTP as used", zap.Error(err), zap.String("otp_id", otp.ID.String()))
		// Continue anyway
	}

	// 4. Find user
	user, err := s.repo.User.FindByEmail(ctx, req.Email)
	if err != nil || user == nil {
		s.log.Error("User not found for verification", zap.Error(err), zap.String("email", req.Email))
		return fmt.Errorf("user not found")
	}

	// 5. Update user verification status
	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	if err := s.repo.User.Update(ctx, user); err != nil {
		s.log.Error("Failed to update user verification", zap.Error(err), zap.String("user_id", user.ID.String()))
		return fmt.Errorf("failed to verify email")
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
		return nil, err
	}

	return session, nil
}

func (s *authService) convertAuthResponse(user *entity.User, session *entity.Session) *response.AuthResponse {
	resp := &response.AuthResponse{
		UserID:     user.ID.String(),
		Email:      user.Email,
		Username:   user.Username,
		Role:       entity.UserRole(user.Role),
		IsVerified: user.EmailVerified,
	}

	if session != nil {
		resp.Token = session.Token.String()
		resp.ExpiresAt = session.ExpiresAt
	}

	return resp
}

func (s *authService) sendVerificationOTP(email string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.SendOTP(ctx, email, string(entity.OTPTypeEmailVerification)); err != nil {
		s.log.Error("Failed to send verification OTP", zap.Error(err), zap.String("email", email))
	}
}
