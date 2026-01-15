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

type MovieService interface {
	// Get movies dengan pagination (sesuai requirement)
	GetMovies(ctx context.Context, req *request.PaginatedRequest, releaseStatus *string) (*response.PaginatedResponse[response.MovieResponse], error)

	// Get movie detail
	GetMovieByID(ctx context.Context, movieID string) (*response.MovieDetailResponse, error)

	// Admin operations (optional)
	CreateMovie(ctx context.Context, req *request.MovieRequest) (*response.MovieResponse, error)
	UpdateMovie(ctx context.Context, movieID string, req *request.MovieUpdateRequest) (*response.MovieResponse, error)
	DeleteMovie(ctx context.Context, movieID string) error
}

type movieService struct {
	repo *repository.Repository // grouping Movie, Genre, & MovieGenre Repositories
	log  *zap.Logger
}

func NewMovieService(
	repo *repository.Repository,
	log *zap.Logger,
) MovieService {
	return &movieService{
		repo: repo,
		log:  log.With(zap.String("service", "movie")),
	}
}

func (s *movieService) GetMovies(ctx context.Context, req *request.PaginatedRequest, releaseStatus *string) (*response.PaginatedResponse[response.MovieResponse], error) {
	limit := req.Limit()
	offset := req.Offset()
	page := req.Page
	perPage := req.PerPage

	// Get movies from repository
	movies, err := s.repo.Movie.FindAll(ctx, offset, limit, releaseStatus)
	if err != nil {
		s.log.Error("Failed to get movies from repository",
			zap.Error(err),
			zap.Int("page", page),
			zap.Int("per_page", perPage),
			zap.Stringp("release_status", releaseStatus),
		)
		return nil, fmt.Errorf("failed to get movies")
	}

	// Get total count
	total, err := s.repo.Movie.CountAll(ctx, releaseStatus)
	if err != nil {
		s.log.Error("Failed to count movies",
			zap.Error(err),
			zap.Stringp("release_status", releaseStatus),
		)
		return nil, fmt.Errorf("failed to count movies")
	}

	// Convert to response
	movieResponses := make([]response.MovieResponse, len(movies))
	for i, movie := range movies {
		genres, err := s.repo.Genre.FindByMovieID(ctx, movie.ID)
		if err != nil {
			s.log.Warn("Failed to get genres for movie",
				zap.Error(err),
				zap.String("movie_id", movie.ID.String()),
			)
		}

		genreNames := make([]string, len(genres))
		for j, genre := range genres {
			genreNames[j] = genre.Name
		}

		reviewCount := 0 // TODO: get from review repo
		movieResponses[i] = s.convertToMovieResponse(movie, genreNames, reviewCount)
	}

	s.log.Info("Movies retrieved",
		zap.Int("count", len(movies)),
		zap.Int64("total", total),
		zap.Int("page", page),
		zap.Int("per_page", perPage),
	)

	return response.NewPaginatedResponse(movieResponses, page, perPage, total), nil
}

func (s *movieService) GetMovieByID(ctx context.Context, movieID string) (*response.MovieDetailResponse, error) {
	// Parse movie ID
	id, err := uuid.Parse(movieID)
	if err != nil {
		s.log.Warn("Invalid movie ID format",
			zap.String("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid movie ID")
	}

	// Get movie
	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get movie by ID",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		return nil, fmt.Errorf("failed to get movie")
	}

	if movie == nil {
		return nil, fmt.Errorf("movie not found")
	}

	// Get genres
	genres, err := s.repo.Genre.FindByMovieID(ctx, movie.ID)
	if err != nil {
		s.log.Warn("Failed to get genres for movie",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		// Continue with empty genres
	}

	// Get genre names
	genreNames := make([]string, len(genres))
	for i, genre := range genres {
		genreNames[i] = genre.Name
	}

	// TODO: Get review count
	reviewCount := 0

	s.log.Info("Movie retrieved",
		zap.String("movie_id", movieID),
		zap.String("title", movie.Title),
	)

	return s.convertToMovieDetailResponse(movie, genreNames, reviewCount), nil
}

// CreateMovie - admin only (optional)
func (s *movieService) CreateMovie(ctx context.Context, req *request.MovieRequest) (*response.MovieResponse, error) {
	// Validate request
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Create movie validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	// Parse release date
	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		s.log.Warn("Invalid release date format",
			zap.String("release_date", req.ReleaseDate),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid release date format")
	}

	// Parse release status
	var releaseStatus entity.ReleaseStatus
	switch req.ReleaseStatus {
	case "now_playing":
		releaseStatus = entity.ReleaseStatusNowPlaying
	case "coming_soon":
		releaseStatus = entity.ReleaseStatusComingSoon
	default:
		return nil, fmt.Errorf("invalid release status")
	}

	// Check if genres exist
	genreUUIDs := make([]uuid.UUID, 0, len(req.GenreIDs))
	for _, genreIDStr := range req.GenreIDs {
		genreID, err := uuid.Parse(genreIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid genre ID: %s", genreIDStr)
		}

		genre, err := s.repo.Genre.FindByID(ctx, genreID)
		if err != nil {
			s.log.Error("Failed to check genre existence",
				zap.Error(err),
				zap.String("genre_id", genreIDStr),
			)
			return nil, fmt.Errorf("failed to check genre")
		}
		if genre == nil {
			return nil, fmt.Errorf("genre not found: %s", genreIDStr)
		}

		genreUUIDs = append(genreUUIDs, genreID)
	}

	// Create movie entity
	now := time.Now()
	movie := &entity.Movie{
		Base: entity.Base{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:             req.Title,
		Description:       req.Description,
		PosterURL:         req.PosterURL,
		Rating:            0.0, // Default rating
		ReleaseDate:       releaseDate,
		DurationInMinutes: req.DurationInMinutes,
		ReleaseStatus:     releaseStatus,
	}

	// Save movie
	if err := s.repo.Movie.Create(ctx, movie); err != nil {
		s.log.Error("Failed to create movie",
			zap.Error(err),
			zap.String("title", req.Title),
		)
		return nil, fmt.Errorf("failed to create movie")
	}

	// Create movie-genre relationships
	if len(genreUUIDs) > 0 {
		movieGenres := make([]*entity.MovieGenre, len(genreUUIDs))
		for i, genreID := range genreUUIDs {
			movieGenres[i] = &entity.MovieGenre{
				BaseSimple: entity.BaseSimple{
					ID:        uuid.New(),
					CreatedAt: now,
				},
				MovieID: movie.ID,
				GenreID: genreID,
			}
		}

		if err := s.repo.MovieGenre.CreateBatch(ctx, movieGenres); err != nil {
			s.log.Error("Failed to create movie-genre relationships",
				zap.Error(err),
				zap.String("movie_id", movie.ID.String()),
			)
			// Rollback movie creation? For now, just log error
		}
	}

	// Get genre names for response
	genreNames := make([]string, len(genreUUIDs))
	for i, genreID := range genreUUIDs {
		genre, _ := s.repo.Genre.FindByID(ctx, genreID)
		if genre != nil {
			genreNames[i] = genre.Name
		}
	}

	s.log.Info("Movie created",
		zap.String("movie_id", movie.ID.String()),
		zap.String("title", movie.Title),
		zap.Int("genre_count", len(genreUUIDs)),
	)

	movieResp := s.convertToMovieResponse(movie, genreNames, 0)
	return &movieResp, nil
}

// UpdateMovie - admin only (optional)
func (s *movieService) UpdateMovie(ctx context.Context, movieID string, req *request.MovieUpdateRequest) (*response.MovieResponse, error) {
	// Parse movie ID
	id, err := uuid.Parse(movieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie ID")
	}

	// Get existing movie
	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil || movie == nil {
		return nil, fmt.Errorf("movie not found")
	}

	// Update fields if provided
	updated := false

	if req.Title != nil && *req.Title != movie.Title {
		movie.Title = *req.Title
		updated = true
	}

	if req.Description != nil {
		movie.Description = req.Description
		updated = true
	}

	if req.PosterURL != nil {
		movie.PosterURL = req.PosterURL
		updated = true
	}

	if req.ReleaseDate != nil {
		releaseDate, err := time.Parse("2006-01-02", *req.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("invalid release date format")
		}
		movie.ReleaseDate = releaseDate
		updated = true
	}

	if req.DurationInMinutes != nil && *req.DurationInMinutes != movie.DurationInMinutes {
		movie.DurationInMinutes = *req.DurationInMinutes
		updated = true
	}

	if req.ReleaseStatus != nil {
		var releaseStatus entity.ReleaseStatus
		switch *req.ReleaseStatus {
		case "now_playing":
			releaseStatus = entity.ReleaseStatusNowPlaying
		case "coming_soon":
			releaseStatus = entity.ReleaseStatusComingSoon
		default:
			return nil, fmt.Errorf("invalid release status")
		}
		movie.ReleaseStatus = releaseStatus
		updated = true
	}

	if updated {
		movie.UpdatedAt = time.Now()
		if err := s.repo.Movie.Update(ctx, movie); err != nil {
			s.log.Error("Failed to update movie",
				zap.Error(err),
				zap.String("movie_id", movieID),
			)
			return nil, fmt.Errorf("failed to update movie")
		}
	}

	// Get genres for response
	genres, _ := s.repo.Genre.FindByMovieID(ctx, movie.ID)
	genreNames := make([]string, len(genres))
	for i, genre := range genres {
		genreNames[i] = genre.Name
	}

	s.log.Info("Movie updated",
		zap.String("movie_id", movieID),
		zap.String("title", movie.Title),
		zap.Bool("was_updated", updated),
	)

	movieResp := s.convertToMovieResponse(movie, genreNames, 0)
	return &movieResp, nil
}

// DeleteMovie - admin only (optional)
func (s *movieService) DeleteMovie(ctx context.Context, movieID string) error {
	// Parse movie ID
	id, err := uuid.Parse(movieID)
	if err != nil {
		return fmt.Errorf("invalid movie ID")
	}

	// Get movie first for logging
	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to find movie")
	}
	if movie == nil {
		return fmt.Errorf("movie not found")
	}

	// Delete movie-genre relationships first
	if err := s.repo.MovieGenre.DeleteByMovieID(ctx, id); err != nil {
		s.log.Warn("Failed to delete movie-genre relationships",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		// Continue anyway
	}

	// Soft delete movie
	if err := s.repo.Movie.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete movie",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		return fmt.Errorf("failed to delete movie")
	}

	s.log.Info("Movie deleted",
		zap.String("movie_id", movieID),
		zap.String("title", movie.Title),
	)

	return nil
}

// Helper methods
func (s *movieService) convertToMovieResponse(movie *entity.Movie, genres []string, reviewCount int) response.MovieResponse {
	// Return struct value
	return response.MovieResponse{
		ID:                movie.ID.String(),
		Title:             movie.Title,
		Description:       movie.Description,
		PosterURL:         movie.PosterURL,
		Rating:            movie.Rating,
		ReviewCount:       reviewCount,
		ReleaseDate:       movie.ReleaseDate.Format("2006-01-02"),
		DurationInMinutes: fmt.Sprintf("%d", movie.DurationInMinutes),
		Genres:            genres,
		ReleaseStatus:     s.formatReleaseStatus(movie.ReleaseStatus),
		CreatedAt:         movie.CreatedAt,
	}
}

func (s *movieService) formatReleaseStatus(status entity.ReleaseStatus) string {
	switch status {
	case entity.ReleaseStatusNowPlaying:
		return "now"
	case entity.ReleaseStatusComingSoon:
		return "coming_soon"
	default:
		return string(status)
	}
}

func (s *movieService) convertToMovieDetailResponse(movie *entity.Movie, genres []string, reviewCount int) *response.MovieDetailResponse {
	movieResp := s.convertToMovieResponse(movie, genres, reviewCount)
	return &response.MovieDetailResponse{
		MovieResponse: movieResp,
		Description:   movie.Description,
		UpdatedAt:     &movie.UpdatedAt,
	}
}
