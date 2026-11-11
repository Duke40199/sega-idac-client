package port

import (
	"context"

	"github.com/Duke40199/sega-idac-client/internal/domain/entity"
)

type AllNetClient interface {
	GetListStore(ctx context.Context, gameCode, languageCode, areaCode string) ([]entity.StoreLocation, string, error)
}

// SegaIDACClient talks to SEGA's Initial D The Arcade public ranking endpoints.
// Every method takes a context so callers (e.g. gRPC handlers) can propagate
// cancellation and deadlines down to the outbound HTTP request.
type SegaIDACClient interface {
	GetListTimeTrail(ctx context.Context, courseID, areaCode, carID, spec string) ([]entity.TimeAttackRecord, error)
	GetTeamRanking(ctx context.Context, roundNum int, rankCode string) ([]entity.TeamRecord, error)
	GetListOBRanking(ctx context.Context, roundNum, areaCode string) (*entity.IdacOBRankingResponse, error)
	GetOBRankingByIGN(ctx context.Context, ign, roundNum, areaCode string) (*entity.OBRankingRecord, error)
	GetCurrentRound(ctx context.Context) (int, error)
	FetchConst(ctx context.Context) (*entity.IdacConstResponse, error)
	GetListPlayerGrade(ctx context.Context, areaCode string) (*entity.IdacPlayerRankingResponse, error)
	GetPlayerGradeByIGN(ctx context.Context, ign, areaCode string) (*entity.PlayerRankingRecord, error)
}
