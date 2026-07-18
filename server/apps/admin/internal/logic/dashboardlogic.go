package logic

import (
	"context"
	"fmt"
	"server/apps/admin/internal/svc"
	"server/apps/admin/internal/types"
	"server/common/errorx"
	"time"
)

type DashboardLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardLogic {
	return &DashboardLogic{ctx: ctx, svcCtx: svcCtx}
}

// Dashboard 控制台概览：KPI、流量趋势、近期操作
func (l *DashboardLogic) Dashboard() (resp *types.DashboardResp, err error) {
	m := l.svcCtx.Models

	activeLinks, err := m.ShortLink.CountWhere(l.ctx, 1)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	totalVisits, err := m.ShortLink.SumClicks(l.ctx)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	activeTokens, err := m.ApiKey.CountWhere(l.ctx, 1)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	blocked, err := m.DomainBlacklist.Count(l.ctx)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}

	resp = &types.DashboardResp{
		Kpis: []types.KpiItem{
			{Key: "links", Label: "Total Active Links", Value: formatInt(activeLinks), Badge: ""},
			{Key: "visits", Label: "Total Visits", Value: formatInt(totalVisits), Badge: ""},
			{Key: "tokens", Label: "Active API Tokens", Value: formatInt(activeTokens), Badge: ""},
			{Key: "blocked", Label: "Blocked Domains", Value: formatInt(blocked), Badge: ""},
		},
	}

	// 流量趋势（近 7 天）
	traffic, err := m.ActionLog.CountByDay(l.ctx, 7)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	for i := 6; i >= 0; i-- {
		day := todayMinus(i)
		resp.Traffic = append(resp.Traffic, types.TrafficPoint{
			Date:  day,
			Value: traffic[day],
		})
	}

	// 近期操作：合并最近短链与黑名单
	links, _, err := m.ShortLink.FindPageWithUser(l.ctx, 1, 3)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	for _, it := range links {
		resp.Actions = append(resp.Actions, types.AdminActionItem{
			Title: "New short link created",
			Meta:  "CODE: " + it.Code,
			Time:  it.CreatedAt,
		})
	}
	bls, _, err := m.DomainBlacklist.FindPage(l.ctx, 1, 2)
	if err != nil {
		return nil, errorx.Internal(err.Error())
	}
	for _, it := range bls {
		resp.Actions = append(resp.Actions, types.AdminActionItem{
			Title: "Domain blacklisted",
			Meta:  "DOMAIN: " + it.Domain,
			Time:  it.CreatedAt,
		})
	}

	return resp, nil
}

func formatInt(v int64) string {
	return fmt.Sprintf("%d", v)
}

// todayMinus 返回 n 天前的日期（格式 2006-01-02，与 MySQL date() 一致）
func todayMinus(n int) string {
	return time.Now().AddDate(0, 0, -n).Format("2006-01-02")
}
