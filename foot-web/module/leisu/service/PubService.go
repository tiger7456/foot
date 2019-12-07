package service

import (
	"encoding/json"
	"strings"
	"tesou.io/platform/foot-parent/foot-api/common/base"
	"tesou.io/platform/foot-parent/foot-api/module/analy/pojo"
	"tesou.io/platform/foot-parent/foot-core/module/analy/service"
	"tesou.io/platform/foot-parent/foot-web/module/leisu/constants"
	"tesou.io/platform/foot-parent/foot-web/module/leisu/utils"
	"tesou.io/platform/foot-parent/foot-web/module/leisu/vo"
	"time"
)

/**
发布推荐
*/
type PubService struct {
	service.AnalyService
	MatchPoolService
	PubLimitService
}

const content = "个人业余开发的一款足球分析程序，计算分析得出的结果.目前处于验证阶段，会持续同步发布出来，如发现近期命中率可以，可以选择跟进或参考，但一定要慎跟！！！！个人业余开发的一款足球分析程序，计算分析得出的结果.目前处于验证阶段，会持续同步发布出来，如发现近期命中率可以，可以选择跟进或参考，但一定要慎跟！！！"

/**
发布北京单场胜负过关
*/
func (this *PubService) PubBJDC() {
	mainTeam := true
	//获取分析计算出的比赛列表
	analyList := this.AnalyService.GetPubDataList("Euro81_616Service", mainTeam)
	if len(analyList) <= 0 {
		base.Log.Info("当前无可发布的比赛!!!!")
		return
	}
	//获取发布池的比赛列表
	matchPool := this.MatchPoolService.GetMatchList()
	//适配比赛,获取发布列表
	pubList := make(map[*pojo.AnalyResult]*vo.MatchVO, 0)
	for _, analy := range analyList {
		analy_mainTeam := analy.MainTeamId
		for _, match := range matchPool {
			if strings.EqualFold(analy_mainTeam, match.MainTeam) {
				pubList[analy] = match
				break;
			}
		}
	}
	//发布比赛
	for analy, match := range pubList {
		//检查是否还有发布次数
		count := this.PubLimitService.HasPubCount()
		if !count {
			base.Log.Error("已经没有发布次数啦,请更换cookies帐号!")
			break;
		}
		//检查是否是重复发布

		//----
		action := this.BJDCAction(match, mainTeam)
		if nil != action && action.Code == 0 && action.Id > 0 {
			analy.LeisuPubd = true
			this.AnalyService.Modify(analy)
		}
		//需要间隔6分钟，再进行下一次发布
		time.Sleep(6 * time.Minute)
	}
}

/**
发布比赛
*/
func (this *PubService) BJDCAction(param *vo.MatchVO, mainTeam bool) *vo.PubRespVO {
	current_date := time.Now().Format("20060102150405")
	pubVO := new(vo.PubVO)
	pubVO.Title = "命中率验证中:" + current_date + "+" + param.MainTeam + "VS" + param.GuestTeam
	pubVO.Content = content
	pubVO.Multiple = 0
	pubVO.Price = 0
	//设置赔率
	oddData := param.GetBJDCOddData(mainTeam)
	if oddData == nil {
		//没有找到北单胜负过关的选项
		return nil;
	}
	infvo := vo.MatchINFVO{}
	infvo.Id = param.DataId
	infvo.Selects = []int{oddData.DataSelects}
	infvo.Values = []float64{oddData.DataOdd}
	pubVO.Data = []vo.MatchINFVO{infvo}
	//执行发布
	post := this.PubPost(pubVO)
	base.Log.Info("发布结果:" + pubVO.Title + " " + post.ToString())
	return post
}

/**
处理http post
*/
func (this *PubService) PubPost(param *vo.PubVO) *vo.PubRespVO {
	data := utils.Post(constants.PUB_URL, param)
	if len(data) <= 0 {
		base.Log.Error("PubPost:获取到的数据为空")
		return nil
	}
	base.Log.Info("http post 请求返回:" + data)
	resp := new(vo.PubRespVO)
	json.Unmarshal([]byte(data), resp)
	return resp
}