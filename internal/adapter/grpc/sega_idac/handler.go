// Package segaidac contains the gRPC adapter that exposes the SEGA IDAC domain
// client (port.SegaIDACClient) over the SegaIDACService gRPC API. It is a thin
// translation layer: it maps protobuf requests to client calls and domain
// entities back to protobuf responses, and performs no business logic of its own.
package segaidac

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	segaidacv1 "github.com/Duke40199/sega-idac-client/gen/segaidacv1"
	"github.com/Duke40199/sega-idac-client/internal/domain/entity"
	"github.com/Duke40199/sega-idac-client/internal/domain/port"
)

// Handler implements segaidacv1.SegaIDACServiceServer by delegating to the
// domain SegaIDACClient.
type Handler struct {
	segaidacv1.UnimplementedSegaIDACServiceServer

	client port.SegaIDACClient
}

// NewHandler creates a new SegaIDACService gRPC handler backed by client.
func NewHandler(client port.SegaIDACClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetListTimeTrail(ctx context.Context, req *segaidacv1.GetListTimeTrailRequest) (*segaidacv1.GetListTimeTrailResponse, error) {
	records, err := h.client.GetListTimeTrail(ctx, req.GetCourseId(), req.GetAreaCode(), req.GetCarId(), req.GetSpec())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get list time trail: %v", err)
	}
	return &segaidacv1.GetListTimeTrailResponse{
		Records: toProtoTimeAttackRecords(records),
	}, nil
}

func (h *Handler) GetTeamRanking(ctx context.Context, req *segaidacv1.GetTeamRankingRequest) (*segaidacv1.GetTeamRankingResponse, error) {
	records, err := h.client.GetTeamRanking(ctx, int(req.GetRoundNum()), req.GetRankCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get team ranking: %v", err)
	}
	return &segaidacv1.GetTeamRankingResponse{
		Records: toProtoTeamRecords(records),
	}, nil
}

func (h *Handler) GetListOBRanking(ctx context.Context, req *segaidacv1.GetListOBRankingRequest) (*segaidacv1.GetListOBRankingResponse, error) {
	resp, err := h.client.GetListOBRanking(ctx, req.GetRoundNum(), req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get list ob ranking: %v", err)
	}
	return &segaidacv1.GetListOBRankingResponse{
		CalcDate: resp.CalcDate,
		Records:  toProtoOBRankingRecords(resp.Records),
	}, nil
}

func (h *Handler) GetOBRankingByIGN(ctx context.Context, req *segaidacv1.GetOBRankingByIGNRequest) (*segaidacv1.GetOBRankingByIGNResponse, error) {
	record, err := h.client.GetOBRankingByIGN(ctx, req.GetIgn(), req.GetRoundNum(), req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get ob ranking by ign: %v", err)
	}
	if record == nil {
		return &segaidacv1.GetOBRankingByIGNResponse{Found: false}, nil
	}
	return &segaidacv1.GetOBRankingByIGNResponse{
		Found:  true,
		Record: toProtoOBRankingRecord(record),
	}, nil
}

func (h *Handler) GetCurrentRound(ctx context.Context, _ *segaidacv1.GetCurrentRoundRequest) (*segaidacv1.GetCurrentRoundResponse, error) {
	round, err := h.client.GetCurrentRound(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get current round: %v", err)
	}
	return &segaidacv1.GetCurrentRoundResponse{Round: int32(round)}, nil
}

func (h *Handler) FetchConst(ctx context.Context, _ *segaidacv1.FetchConstRequest) (*segaidacv1.FetchConstResponse, error) {
	resp, err := h.client.FetchConst(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetch const: %v", err)
	}
	return &segaidacv1.FetchConstResponse{
		Cars:   toProtoCarMetadataList(resp.Cars),
		Styles: toProtoCarStyleMetadataList(resp.Styles),
	}, nil
}

func (h *Handler) GetListPlayerGrade(ctx context.Context, req *segaidacv1.GetListPlayerGradeRequest) (*segaidacv1.GetListPlayerGradeResponse, error) {
	resp, err := h.client.GetListPlayerGrade(ctx, req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get list player grade: %v", err)
	}
	return &segaidacv1.GetListPlayerGradeResponse{
		CalcDate: resp.CalcDate,
		Records:  toProtoPlayerRankingRecords(resp.Records),
	}, nil
}

func (h *Handler) GetPlayerGradeByIGN(ctx context.Context, req *segaidacv1.GetPlayerGradeByIGNRequest) (*segaidacv1.GetPlayerGradeByIGNResponse, error) {
	record, err := h.client.GetPlayerGradeByIGN(ctx, req.GetIgn(), req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get player grade by ign: %v", err)
	}
	if record == nil {
		return &segaidacv1.GetPlayerGradeByIGNResponse{Found: false}, nil
	}
	return &segaidacv1.GetPlayerGradeByIGNResponse{
		Found:  true,
		Record: toProtoPlayerRankingRecord(record),
	}, nil
}

// ----- entity -> proto mappers -----

func toProtoTimeAttackRecords(in []entity.TimeAttackRecord) []*segaidacv1.TimeAttackRecord {
	out := make([]*segaidacv1.TimeAttackRecord, 0, len(in))
	for _, r := range in {
		out = append(out, &segaidacv1.TimeAttackRecord{
			Rank:       r.Rank,
			Name:       r.Name,
			ShopName:   r.ShopName,
			Record:     r.Record,
			CarName:    r.CarName,
			MyTitleId:  r.MyTitleID,
			UpdateDate: r.UpdateDate,
		})
	}
	return out
}

func toProtoTeamRecords(in []entity.TeamRecord) []*segaidacv1.TeamRecord {
	out := make([]*segaidacv1.TeamRecord, 0, len(in))
	for _, r := range in {
		out = append(out, &segaidacv1.TeamRecord{
			Rank:           r.Rank,
			TeamName:       r.TeamName,
			ShopName:       r.ShopName,
			Country:        int32(r.Country),
			AceUserName:    r.AceUserName,
			LeaderUserName: r.LeaderUserName,
			Point:          r.Point,
			LeagueEmoji:    r.LeagueEmoji,
		})
	}
	return out
}

func toProtoOBRankingRecord(r *entity.OBRankingRecord) *segaidacv1.OBRankingRecord {
	if r == nil {
		return nil
	}
	return &segaidacv1.OBRankingRecord{
		Rank:               r.Rank,
		Name:               r.Name,
		ShopName:           r.ShopName,
		UpdateDate:         r.UpdateDate,
		CarName:            r.CarName,
		Point:              r.Point,
		PrideId:            r.PrideId,
		PridePoint:         int32(r.PridePoint),
		OnlineBattleRankId: r.OnlineBattleRankId,
		StarCnt:            int32(r.StarCnt),
	}
}

func toProtoOBRankingRecords(in []entity.OBRankingRecord) []*segaidacv1.OBRankingRecord {
	out := make([]*segaidacv1.OBRankingRecord, 0, len(in))
	for i := range in {
		out = append(out, toProtoOBRankingRecord(&in[i]))
	}
	return out
}

func toProtoPlayerRankingRecord(r *entity.PlayerRankingRecord) *segaidacv1.PlayerRankingRecord {
	if r == nil {
		return nil
	}
	return &segaidacv1.PlayerRankingRecord{
		Id:         int32(r.ID),
		Rank:       int32(r.Rank),
		Name:       r.Name,
		ShopName:   r.ShopName,
		GradeId:    r.GradeID,
		GradeExp:   int32(r.GradeExp),
		MyTitleId:  r.MytitleID,
		UpdateDate: r.UpdateDate,
		NumberIcon: r.NumberIcon,
	}
}

func toProtoPlayerRankingRecords(in []entity.PlayerRankingRecord) []*segaidacv1.PlayerRankingRecord {
	out := make([]*segaidacv1.PlayerRankingRecord, 0, len(in))
	for i := range in {
		out = append(out, toProtoPlayerRankingRecord(&in[i]))
	}
	return out
}

func toProtoCarMetadataList(in []entity.CarMetadata) []*segaidacv1.CarMetadata {
	out := make([]*segaidacv1.CarMetadata, 0, len(in))
	for _, c := range in {
		out = append(out, &segaidacv1.CarMetadata{
			Id:                  c.ID,
			SegaCarId:           c.SegaCarID,
			Name:                c.Name,
			ModelCode:           c.ModelCode,
			Maker:               c.Maker,
			BaseStyleName:       c.BaseStyleName,
			NormalizedBaseStyle: c.NormalizedBaseStyle,
			CarStyleIds:         c.CarStyleIDs,
			SpecIds:             c.SpecIDs,
			SpecNames:           c.SpecNames,
		})
	}
	return out
}

func toProtoCarStyleMetadataList(in []entity.CarStyleMetadata) []*segaidacv1.CarStyleMetadata {
	out := make([]*segaidacv1.CarStyleMetadata, 0, len(in))
	for _, s := range in {
		out = append(out, &segaidacv1.CarStyleMetadata{
			Id:             s.ID,
			StyleCarId:     s.StyleCarID,
			CarId:          s.CarID,
			Name:           s.Name,
			RouteStyleName: s.RouteStyleName,
		})
	}
	return out
}
