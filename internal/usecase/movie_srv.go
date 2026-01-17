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
	GetMovies(ctx context.Context, req *request.PaginatedRequest, releaseStatus *string) (*response.PaginatedResponse[response.MovieResponse], error)
	GetMovieByID(ctx context.Context, movieID string) (*response.MovieDetailResponse, error)
	CreateMovie(ctx context.Context, req *request.MovieRequest) (*response.MovieResponse, error)
	UpdateMovie(ctx context.Context, movieID string, req *request.MovieUpdateRequest) (*response.MovieResponse, error)
	DeleteMovie(ctx context.Context, movieID string) error
}

type movieService struct {
	repo *repository.Repository
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

	// Get movies with pagination and filter
	movies, err := s.repo.Movie.FindAll(ctx, limit, offset, releaseStatus)
	if err != nil {
		s.log.Error("Failed to get movies",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("per_page", req.PerPage),
			zap.Stringp("release_status", releaseStatus),
		)
		return nil, fmt.Errorf("get movies: %w", err)
	}

	// Get total count for pagination metadata
	total, err := s.repo.Movie.CountAll(ctx, releaseStatus)
	if err != nil {
		s.log.Error("Failed to count movies",
			zap.Error(err),
			zap.Stringp("release_status", releaseStatus),
		)
		return nil, fmt.Errorf("count movies: %w", err)
	}

	// Convert each movie to response with additional data
	movieResponses := make([]response.MovieResponse, len(movies))
	for i, movie := range movies {
		// Get associated genres
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

		// Get review statistics
		avgRating, reviewCount, err := s.repo.Review.GetMovieReviewStats(ctx, movie.ID)
		if err != nil {
			// Log error but continue
			s.log.Warn("Failed to get review stats for movie",
				zap.Error(err),
				zap.String("movie_id", movie.ID.String()),
			)
			// Use default values
			avgRating = movie.Rating
			reviewCount = 0
		} else if avgRating > 0 { // Update movie rating if reviews exist
			movie.Rating = avgRating
		}

		movieResponses[i] = response.MovieToResponse(movie, genreNames, int(reviewCount))
	}

	s.log.Info("Movies retrieved",
		zap.Int("count", len(movies)),
		zap.Int64("total", total),
		zap.Int("page", req.Page),
		zap.Int("per_page", req.PerPage),
	)

	return response.NewPaginatedResponse(movieResponses, req.Page, req.PerPage, total), nil
}

func (s *movieService) GetMovieByID(ctx context.Context, movieID string) (*response.MovieDetailResponse, error) {
	id, err := uuid.Parse(movieID)
	if err != nil {
		s.log.Warn("Invalid movie ID format",
			zap.String("movie_id", movieID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid movie id: %w", err)
	}

	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get movie by ID",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		return nil, fmt.Errorf("get movie by id: %w", err)
	}

	if movie == nil {
		return nil, fmt.Errorf("movie not found")
	}

	genres, err := s.repo.Genre.FindByMovieID(ctx, movie.ID)
	if err != nil {
		s.log.Warn("Failed to get genres for movie",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
	}

	genreNames := make([]string, len(genres))
	for i, genre := range genres {
		genreNames[i] = genre.Name
	}

	avgRating, reviewCount, err := s.repo.Review.GetMovieReviewStats(ctx, movie.ID)
	if err != nil {
		s.log.Warn("Failed to get review stats for movie",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		// Use default values
		reviewCount = 0
	} else if avgRating > 0 {
		// Update movie rating from reviews
		movie.Rating = avgRating
	}

	s.log.Info("Movie retrieved",
		zap.String("movie_id", movieID),
		zap.String("title", movie.Title),
		zap.Int64("review_count", reviewCount),
		zap.Float64("avg_rating", avgRating),
	)

	detailMovie := response.MovieToDetailResponse(movie, genreNames, int(reviewCount))
	return &detailMovie, nil
}

func (s *movieService) CreateMovie(ctx context.Context, req *request.MovieRequest) (*response.MovieResponse, error) {
	// Validate request data
	if errs := utils.ValidateStruct(req); len(errs) > 0 {
		s.log.Warn("Create movie validation failed", zap.Any("errors", errs))
		return nil, fmt.Errorf("validation failed: %s", utils.FormatValidationErrors(errs))
	}

	releaseDate, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		s.log.Warn("Invalid release date format",
			zap.String("release_date", req.ReleaseDate),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid release date: %w", err)
	}

	/// Validate release status enum
	var releaseStatus entity.ReleaseStatus
	switch req.ReleaseStatus {
	case "now_playing":
		releaseStatus = entity.ReleaseStatusNowPlaying
	case "coming_soon":
		releaseStatus = entity.ReleaseStatusComingSoon
	default:
		return nil, fmt.Errorf("invalid release status: %s", req.ReleaseStatus)
	}

	// Validate genres
	genreUUIDs := make([]uuid.UUID, 0, len(req.GenreIDs))
	for _, genreIDStr := range req.GenreIDs {
		genreID, err := uuid.Parse(genreIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid genre id: %w", err)
		}

		genre, err := s.repo.Genre.FindByID(ctx, genreID)
		if err != nil {
			s.log.Error("Failed to check genre existence",
				zap.Error(err),
				zap.String("genre_id", genreIDStr),
			)
			return nil, fmt.Errorf("check genre: %w", err)
		}
		if genre == nil {
			return nil, fmt.Errorf("genre not found: %s", genreIDStr)
		}

		genreUUIDs = append(genreUUIDs, genreID)
	}

	// Create movie
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
		Rating:            0.0,
		ReleaseDate:       releaseDate,
		DurationInMinutes: req.DurationInMinutes,
		ReleaseStatus:     releaseStatus,
	}

	// Save movie to database
	if err := s.repo.Movie.Create(ctx, movie); err != nil {
		s.log.Error("Failed to create movie",
			zap.Error(err),
			zap.String("title", req.Title),
		)
		return nil, fmt.Errorf("create movie: %w", err)
	}

	// Create movie-genre relationships in batch
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

		// Batch insert for performance
		if err := s.repo.MovieGenre.CreateBatch(ctx, movieGenres); err != nil {
			s.log.Error("Failed to create movie-genre relationships",
				zap.Error(err),
				zap.String("movie_id", movie.ID.String()),
			)
			// Rollback: delete movie if genre relationships fail
			s.repo.Movie.Delete(ctx, movie.ID)
			return nil, fmt.Errorf("create movie-genre relationships: %w", err)
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

	movieResp := response.MovieToResponse(movie, genreNames, 0)
	return &movieResp, nil
}

func (s *movieService) UpdateMovie(ctx context.Context, movieID string, req *request.MovieUpdateRequest) (*response.MovieResponse, error) {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return nil, fmt.Errorf("invalid movie id: %w", err)
	}

	// Find existing movie
	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("find movie: %w", err)
	}
	if movie == nil {
		return nil, fmt.Errorf("movie not found")
	}

	// Apply partial updates only for provided fields
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
			return nil, fmt.Errorf("invalid release date: %w", err)
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
			return nil, fmt.Errorf("invalid release status: %s", *req.ReleaseStatus)
		}
		movie.ReleaseStatus = releaseStatus
		updated = true
	}

	// Update timestamp and save only if changes were made
	if updated {
		movie.UpdatedAt = time.Now()
		if err := s.repo.Movie.Update(ctx, movie); err != nil {
			s.log.Error("Failed to update movie",
				zap.Error(err),
				zap.String("movie_id", movieID),
			)
			return nil, fmt.Errorf("update movie: %w", err)
		}
	}

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

	// Return updated movie response
	movieResp := response.MovieToResponse(movie, genreNames, 0)
	return &movieResp, nil
}

func (s *movieService) DeleteMovie(ctx context.Context, movieID string) error {
	id, err := uuid.Parse(movieID)
	if err != nil {
		return fmt.Errorf("invalid movie id: %w", err)
	}

	movie, err := s.repo.Movie.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find movie: %w", err)
	}
	if movie == nil {
		return fmt.Errorf("movie not found")
	}

	if err := s.repo.MovieGenre.DeleteByMovieID(ctx, id); err != nil {
		s.log.Warn("Failed to delete movie-genre relationships",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
	}

	if err := s.repo.Movie.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete movie",
			zap.Error(err),
			zap.String("movie_id", movieID),
		)
		return fmt.Errorf("delete movie: %w", err)
	}

	s.log.Info("Movie deleted",
		zap.String("movie_id", movieID),
		zap.String("title", movie.Title),
	)

	return nil
}
