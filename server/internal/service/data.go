package service

import (
	"context"
	"fmt"
	"log"
	"os"

	pb "lunar-tear/server/gen/proto"
	"lunar-tear/server/internal/store"
	"lunar-tear/server/internal/userdata"

	"google.golang.org/protobuf/types/known/emptypb"
)

// masterDataBinPath / masterDataBaseVersion 可用环境变量覆盖，便于将来
// 多区域（EN/JP）分别配置；不配置时保持原有 EN 默认值。
//
//	LUNAR_MASTERDATA          主数据文件路径（默认 assets/release/20240404193219.bin.e）
//	LUNAR_MASTERDATA_VERSION  对外返回的版本号前缀（默认 20240404193219）
//
// 文件 mtime 会被拼接进版本号，以便管理员替换主数据后客户端自动失效缓存。
var (
	masterDataBinPath     = envOr("LUNAR_MASTERDATA", "assets/release/20240404193219.bin.e")
	masterDataBaseVersion = envOr("LUNAR_MASTERDATA_VERSION", "20240404193219")
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

type DataServiceServer struct {
	pb.UnimplementedDataServiceServer
	users    store.UserRepository
	sessions store.SessionRepository
}

func NewDataServiceServer(users store.UserRepository, sessions store.SessionRepository) *DataServiceServer {
	return &DataServiceServer{users: users, sessions: sessions}
}

func (s *DataServiceServer) GetLatestMasterDataVersion(ctx context.Context, _ *emptypb.Empty) (*pb.MasterDataGetLatestVersionResponse, error) {
	version := masterDataBaseVersion
	if info, err := os.Stat(masterDataBinPath); err == nil {
		version = fmt.Sprintf("%s_%d", masterDataBaseVersion, info.ModTime().UnixMilli())
	} else {
		log.Printf("[DataService] stat %s: %v (falling back to base version)", masterDataBinPath, err)
	}
	log.Printf("[DataService] GetLatestMasterDataVersion -> %s", version)
	return &pb.MasterDataGetLatestVersionResponse{
		LatestMasterDataVersion: version,
	}, nil
}

func (s *DataServiceServer) GetUserDataNameV2(ctx context.Context, _ *emptypb.Empty) (*pb.UserDataGetNameResponseV2, error) {
	log.Printf("[DataService] GetUserDataNameV2")
	return &pb.UserDataGetNameResponseV2{
		TableNameList: []*pb.TableNameList{
			{TableName: defaultTableNames()},
		},
	}, nil
}

func (s *DataServiceServer) GetUserData(ctx context.Context, req *pb.UserDataGetRequest) (*pb.UserDataGetResponse, error) {
	log.Printf("[DataService] GetUserData: tables=%v", req.TableName)

	userId := CurrentUserId(ctx, s.users, s.sessions)
	user, err := s.users.LoadUser(userId)
	if err != nil {
		return nil, fmt.Errorf("snapshot user: %w", err)
	}

	defaults := userdata.FullClientTableMap(user)
	result := userdata.SelectTables(defaults, req.TableName)
	return &pb.UserDataGetResponse{
		UserDataJson: result,
	}, nil
}

func defaultTableNames() []string {
	return []string{
		"IUser",
		"IUserApple",
		"IUserAutoSaleSettingDetail",
		"IUserBeginnerCampaign",
		"IUserBigHuntMaxScore",
		"IUserBigHuntProgressStatus",
		"IUserBigHuntScheduleMaxScore",
		"IUserBigHuntStatus",
		"IUserBigHuntWeeklyMaxScore",
		"IUserBigHuntWeeklyStatus",
		"IUserCageOrnamentReward",
		"IUserCharacter",
		"IUserCharacterBoard",
		"IUserCharacterBoardAbility",
		"IUserCharacterBoardCompleteReward",
		"IUserCharacterBoardStatusUp",
		"IUserCharacterCostumeLevelBonus",
		"IUserCharacterRebirth",
		"IUserCharacterViewerField",
		"IUserComebackCampaign",
		"IUserCompanion",
		"IUserConsumableItem",
		"IUserContentsStory",
		"IUserCostume",
		"IUserCostumeActiveSkill",
		"IUserCostumeAwakenStatusUp",
		"IUserCostumeLevelBonusReleaseStatus",
		"IUserCostumeLotteryEffect",
		"IUserCostumeLotteryEffectAbility",
		"IUserCostumeLotteryEffectPending",
		"IUserCostumeLotteryEffectStatusUp",
		"IUserDeck",
		"IUserDeckCharacter",
		"IUserDeckCharacterDressupCostume",
		"IUserDeckLimitContentRestricted",
		"IUserDeckPartsGroup",
		"IUserDeckSubWeaponGroup",
		"IUserDeckTypeNote",
		"IUserDokan",
		"IUserEventQuestDailyGroupCompleteReward",
		"IUserEventQuestGuerrillaFreeOpen",
		"IUserEventQuestLabyrinthSeason",
		"IUserEventQuestLabyrinthStage",
		"IUserEventQuestProgressStatus",
		"IUserEventQuestTowerAccumulationReward",
		"IUserExplore",
		"IUserExploreScore",
		"IUserExtraQuestProgressStatus",
		"IUserFacebook",
		"IUserGem",
		"IUserGimmick",
		"IUserGimmickOrnamentProgress",
		"IUserGimmickSequence",
		"IUserGimmickUnlock",
		"IUserImportantItem",
		"IUserLimitedOpen",
		// "IUserLogin",
		"IUserLoginBonus",
		"IUserMainQuestFlowStatus",
		"IUserMainQuestMainFlowStatus",
		"IUserMainQuestProgressStatus",
		"IUserMainQuestReplayFlowStatus",
		"IUserMainQuestSeasonRoute",
		"IUserMaterial",
		"IUserMission",
		"IUserMissionCompletionProgress",
		"IUserMissionPassPoint",
		"IUserMovie",
		"IUserNaviCutIn",
		"IUserOmikuji",
		"IUserParts",
		"IUserPartsGroupNote",
		"IUserPartsPreset",
		"IUserPartsPresetTag",
		"IUserPartsStatusSub",
		"IUserPortalCageStatus",
		"IUserPossessionAutoConvert",
		"IUserPremiumItem",
		"IUserProfile",
		"IUserPvpDefenseDeck",
		"IUserPvpStatus",
		"IUserPvpWeeklyResult",
		"IUserQuest",
		"IUserQuestAutoOrbit",
		"IUserQuestLimitContentStatus",
		"IUserQuestMission",
		"IUserQuestReplayFlowRewardGroup",
		"IUserQuestSceneChoice",
		"IUserQuestSceneChoiceHistory",
		// "IUserSetting",
		"IUserShopItem",
		"IUserShopReplaceable",
		"IUserShopReplaceableLineup",
		"IUserSideStoryQuest",
		"IUserSideStoryQuestSceneProgressStatus",
		"IUserStatus",
		"IUserThought",
		"IUserTripleDeck",
		"IUserTutorialProgress",
		"IUserWeapon",
		"IUserWeaponAbility",
		"IUserWeaponAwaken",
		"IUserWeaponNote",
		"IUserWeaponSkill",
		"IUserWeaponStory",
		"IUserWebviewPanelMission",
	}
}
