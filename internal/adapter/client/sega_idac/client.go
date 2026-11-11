package sega_idac

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Duke40199/sega-idac-client/internal/config"
	"github.com/Duke40199/sega-idac-client/internal/domain/entity"
	"github.com/Duke40199/sega-idac-client/internal/domain/port"
)

// segaHTTPTimeout bounds every Sega request so a stalled endpoint can't hang a
// gRPC handler or sync goroutine indefinitely. A caller-supplied ctx with an
// earlier deadline still wins.
const segaHTTPTimeout = 15 * time.Second

type SegaIDACClient struct {
	config     config.Config
	httpClient *http.Client
}

func NewSegaIDACClient(cfg config.Config) port.SegaIDACClient {
	return &SegaIDACClient{
		config:     cfg,
		httpClient: &http.Client{Timeout: segaHTTPTimeout},
	}
}

// makeHTTPRequest performs a GET against url honoring ctx and decodes a JSON body into out.
func (c *SegaIDACClient) makeHTTPRequest(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}

// GetListTimeTrail returns the time-attack leaderboard for a course/area/car.
func (c *SegaIDACClient) GetListTimeTrail(ctx context.Context, courseID, areaCode, carID, spec string) ([]entity.TimeAttackRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTimeTrailURLPath
	url = strings.Replace(url, ":courseID", courseID, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	url = strings.Replace(url, ":carID", carID, 1)
	// 2. Fetch + parse
	var data entity.IdacTimeAttackRecordResponse
	if err := c.makeHTTPRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetListTimeTrail: %w", err)
	}
	return data.Records, nil
}

// GetOBRankingByIGN returns the online-battle ranking record matching an in-game name.
func (c *SegaIDACClient) GetOBRankingByIGN(ctx context.Context, ign, roundNum, areaCode string) (*entity.OBRankingRecord, error) {
	normalizedIGN := entity.NormalizeTextWidth(ign)
	listOBRankings, err := c.GetListOBRanking(ctx, roundNum, areaCode)
	if err != nil {
		return nil, err
	}
	if len(listOBRankings.Records) == 0 {
		return nil, nil
	}
	for _, obRanking := range listOBRankings.Records {
		if entity.NormalizeTextWidth(obRanking.Name) == normalizedIGN {
			return &obRanking, nil
		}
	}
	return nil, nil
}

// GetListOBRanking returns the full online-battle ranking for a round/area.
func (c *SegaIDACClient) GetListOBRanking(ctx context.Context, roundNum, areaCode string) (*entity.IdacOBRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListOBRankingURLPath
	url = strings.Replace(url, ":roundNum", roundNum, 1)
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	// 2. Fetch + parse
	var data entity.IdacOBRankingResponse
	if err := c.makeHTTPRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetListOBRanking: %w", err)
	}
	return &data, nil
}

// GetTeamRanking returns the team leaderboard for a round, tagged with league emojis.
func (c *SegaIDACClient) GetTeamRanking(ctx context.Context, roundNum int, rankCode string) ([]entity.TeamRecord, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetTeamRankingUrlPath
	url = strings.Replace(url, ":roundCount", strconv.Itoa(roundNum), 1)
	url = strings.Replace(url, ":rankType", rankCode, 1)
	// 2. Fetch + parse
	var data entity.IdacTeamRankingResponse
	if err := c.makeHTTPRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetTeamRanking: %w", err)
	}
	foundTeams := make([]entity.TeamRecord, 0)
	// set league emoji values for each team
	for _, foundTeam := range data.Records {
		foundTeam.LeagueEmoji = entity.TeamLeagueEmojis[rankCode]
		foundTeams = append(foundTeams, foundTeam)
	}
	return foundTeams, nil
}

// GetCurrentRound returns the currently active tournament round number.
func (c *SegaIDACClient) GetCurrentRound(ctx context.Context) (int, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetCurrentRoundUrlPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return -1, fmt.Errorf("creating request: %w", err)
	}
	respRound, err := c.httpClient.Do(req)
	if err != nil {
		return -1, err
	}
	defer respRound.Body.Close()
	if respRound.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("sega api status: %d", respRound.StatusCode)
	}
	bodyBytes, err := io.ReadAll(respRound.Body)
	if err != nil {
		return -1, fmt.Errorf("error while reading response: %w", err)
	}

	roundStr := strings.TrimSpace(string(bodyBytes))
	roundNum, err := strconv.Atoi(roundStr)
	if err != nil {
		return -1, fmt.Errorf("parsing round number %q: %w", roundStr, err)
	}
	return roundNum, nil
}

// GetListPlayerGrade returns the player-grade ranking for an area.
func (c *SegaIDACClient) GetListPlayerGrade(ctx context.Context, areaCode string) (*entity.IdacPlayerRankingResponse, error) {
	// 1. Build URL
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListPlayerGradeUrlPath
	url = strings.Replace(url, ":areaCode", areaCode, 1)
	// 2. Fetch + parse
	var data entity.IdacPlayerRankingResponse
	if err := c.makeHTTPRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("GetListPlayerGrade: %w", err)
	}
	return &data, nil
}

// GetPlayerGradeByIGN returns the player-grade record matching an in-game name.
func (c *SegaIDACClient) GetPlayerGradeByIGN(ctx context.Context, ign, areaCode string) (*entity.PlayerRankingRecord, error) {
	listAreaGrade, err := c.GetListPlayerGrade(ctx, areaCode)
	if err != nil {
		return nil, err
	}
	normalizedIgn := entity.NormalizeTextWidth(ign)
	if len(listAreaGrade.Records) == 0 {
		return nil, nil
	}
	for _, record := range listAreaGrade.Records {
		if entity.NormalizeTextWidth(record.Name) == normalizedIgn {
			return &record, nil
		}
	}

	return nil, nil
}

// FetchConst returns the IDAC car/style constant metadata.
func (c *SegaIDACClient) FetchConst(ctx context.Context) (*entity.IdacConstResponse, error) {
	url := c.config.GetSegaClientCfg().SegaIDACHost + c.config.GetSegaClientCfg().GetListConstConfigURLPath
	var data entity.IdacConstResponse
	if err := c.makeHTTPRequest(ctx, url, &data); err != nil {
		return nil, fmt.Errorf("FetchConst: %w", err)
	}
	return &data, nil
}
