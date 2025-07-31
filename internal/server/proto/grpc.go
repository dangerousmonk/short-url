package proto

import (
	context "context"
	"errors"

	"github.com/dangerousmonk/short-url/cmd/config"
	"github.com/dangerousmonk/short-url/internal/auth"
	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/dangerousmonk/short-url/internal/models"
	"github.com/dangerousmonk/short-url/internal/service"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// ShortenerGRPCServer Supports all the service methods
type ShortenerGRPCServer struct {
	UnimplementedURLShortenerServiceServer

	service service.URLShortener
	cfg     *config.Config
}

// NewShortenerGRPCServer creates the ShortenerGRPCServer structure and returns a pointer to freshly created struct.
func NewShortenerGRPCServer(s service.URLShortener, cfg *config.Config) *ShortenerGRPCServer {
	return &ShortenerGRPCServer{
		service: s,
		cfg:     cfg,
	}
}

// Create is used to create single short URL.
func (srv *ShortenerGRPCServer) Create(ctx context.Context, req *ShortenRequest) (*ShortenResponse, error) {
	userID := getKeyFromCtx(ctx, string(auth.UserIDHeaderName))
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User id required")
	}

	shortURL, err := srv.service.AddShortURL(ctx, req.Url, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrURLInvalid):
			return nil, status.Error(codes.InvalidArgument, "Invalid URL")

		case errors.Is(err, service.ErrURLExists):
			return nil, status.Error(codes.AlreadyExists, "URL already exists")

		case errors.Is(err, service.ErrSaveFailed):
			return nil, status.Error(codes.Internal, err.Error())
		default:
			logging.Log.Warnf("Shorten unknown err:%v", err)
			return nil, status.Error(codes.Internal, err.Error())
		}

	}
	var resp ShortenResponse
	resp.Result = shortURL
	return &resp, nil
}

// Create batch short URLs
func (srv *ShortenerGRPCServer) AddBatch(ctx context.Context, req *ShortenBatchRequest) (*ShortenBatchResponse, error) {
	userID := getKeyFromCtx(ctx, string(auth.UserIDHeaderName))
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User id required")
	}

	urls := make([]models.APIBatchModel, len(req.Items))
	for i, item := range req.Items {
		urls[i] = models.APIBatchModel{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		}
	}

	result, err := srv.service.AddBatch(ctx, urls, userID)

	if err != nil {
		switch {
		case errors.Is(err, service.ErrTooManyURLs):
			return nil, status.Error(codes.InvalidArgument, err.Error())

		case errors.Is(err, service.ErrNoValidURLs):
			return nil, status.Error(codes.InvalidArgument, err.Error())

		case errors.Is(err, service.ErrHashFailed):
			return nil, status.Error(codes.Internal, err.Error())
		case errors.Is(err, service.ErrSaveBatchFailed):
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	var resp ShortenBatchResponse
	for _, item := range result {
		resp.Items = append(resp.Items, &ShortenBatchResponse_Item{
			CorrelationId: item.CorrelationID, ShortUrl: item.ShortURL})
	}
	return &resp, nil
}

// GetUserURLs returns all URLs saved by user.
func (srv *ShortenerGRPCServer) GetUserURLs(ctx context.Context, req *GetUserURLsRequest) (*GetUserURLsResponse, error) {
	userID := getKeyFromCtx(ctx, string(auth.UserIDHeaderName))
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User id required")
	}

	userURLs, err := srv.service.GetUsersURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if len(userURLs) == 0 {
		return &GetUserURLsResponse{}, nil
	}
	var resp GetUserURLsResponse
	for _, url := range userURLs {
		resp.Urls = append(resp.Urls, &UserURLItem{
			ShortUrl: url.ShortURL, OriginalUrl: url.OriginalURL})
	}
	return &resp, nil
}

// BatchDelete schedules deletion of URLs created by specific user
func (srv *ShortenerGRPCServer) BatchDelete(ctx context.Context, req *DeleteBatchRequest) (*emptypb.Empty, error) {
	userID := getKeyFromCtx(ctx, string(auth.UserIDHeaderName))
	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "User id required")
	}

	if len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "No urls provided to delete")
	}
	srv.service.DeleteBatch(ctx, req.ShortUrls, userID)
	return &emptypb.Empty{}, nil
}

// GetStats retrieves server stats with count of users and saved urls
func (srv *ShortenerGRPCServer) GetServiceStats(ctx context.Context, req *ServiceStatsRequest) (*ServiceStatsResponse, error) {
	stats, err := srv.service.GetStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &ServiceStatsResponse{
		Users: uint32(stats.Users),
		Urls:  uint32(stats.URLs),
	}
	return resp, nil

}

// Ping checks the service health
func (srv ShortenerGRPCServer) Ping(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	err := srv.service.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

// getKeyFromCtx gets value from metadata by given key
func getKeyFromCtx(ctx context.Context, key string) string {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	ids := meta.Get(key)
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}
