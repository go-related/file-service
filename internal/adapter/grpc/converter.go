package grpc

import (
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/go-related/fileservice/proto/pb"
)

func convertPortRequestToDomain(input *pb.PortRequest) *domain.Port {
	if input == nil {
		return nil
	}
	return &domain.Port{
		Id:          input.Id,
		Name:        input.Name,
		City:        input.City,
		Country:     input.Country,
		Alias:       input.Alias,
		Regions:     input.Regions,
		Coordinates: input.Coordinates,
		Province:    input.Province,
		Timezone:    input.Timezone,
		UNLOCs:      input.Unlocs,
		Code:        input.Code,
	}
}
