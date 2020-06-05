package grpchandler

import (
	"errors"
	"io"
	"strconv"

	pb "agungdwiprasetyo.com/backend-microservices/api/proto/storage-service"
	"agungdwiprasetyo.com/backend-microservices/internal/services/storage-service/modules/storage/domain"
	"agungdwiprasetyo.com/backend-microservices/internal/services/storage-service/modules/storage/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// GRPCHandler rpc stream
type GRPCHandler struct {
	uc usecase.StorageUsecase
}

// NewGRPCHandler func
func NewGRPCHandler(uc usecase.StorageUsecase) *GRPCHandler {

	return &GRPCHandler{
		uc: uc,
	}
}

// Register grpc server
func (h *GRPCHandler) Register(server *grpc.Server) {
	pb.RegisterUploadServiceServer(server, h)
}

// Upload method
func (h *GRPCHandler) Upload(stream pb.UploadService_UploadServer) (err error) {

	ctx := stream.Context()
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return grpc.Errorf(codes.Unauthenticated, "missing context metadata")
	}

	fields := meta.Get("filename")
	if len(fields) == 0 {
		return errors.New("missing filename field")
	}
	fileName := fields[0]

	fields = meta.Get("folder")
	if len(fields) == 0 {
		return errors.New("missing folder field")
	}
	folder := fields[0]

	var contentType string
	if u := meta.Get("content_type"); len(u) > 0 {
		contentType = u[0]
	}

	var size int64
	if u := meta.Get("size"); len(u) > 0 {
		s, _ := strconv.Atoi(u[0])
		size = int64(s)
	}

	var buff []byte
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		buff = append(buff, res.Content...)
	}

	res := <-h.uc.Upload(ctx,
		buff,
		&domain.UploadMetadata{
			ContentType: contentType,
			FileSize:    size,
			Filename:    fileName,
		})
	if res.Error != nil {
		return grpc.Errorf(codes.Internal, "%v", res.Error)
	}

	err = stream.SendAndClose(&pb.UploadStatus{
		Message: "Stream file success",
		Code:    pb.UploadStatusCode_Ok,
		File:    "url" + "/" + folder + "/" + fileName,
		Size:    int64(len(buff)),
	})

	return
}