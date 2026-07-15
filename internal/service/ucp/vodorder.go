package ucp

import "context"

func (s *Service) VODOrderMyOrders(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	total, err := s.store.CountVODOrders(ctx, uid, nil)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	rows, err := s.store.VODOrders(ctx, uid, nil, page, pageSize, "id DESC")
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	cost, err := s.store.SumVODOrderCoins(ctx, uid, 1)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	frozen, err := s.store.SumVODOrderCoins(ctx, uid, 0)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	supportCost, err := s.store.SumVODSupportCoins(ctx, uid, false)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	supportFrozen, err := s.store.SumVODSupportCoins(ctx, uid, true)
	if err != nil {
		return nil, -1, "获取求片记录失败", err
	}
	_ = total
	return map[string]interface{}{
		"data":           rows,
		"total_cost":     cost + supportCost,
		"current_frozen": frozen + supportFrozen,
	}, 0, "", nil
}

func (s *Service) VODOrderMySupports(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	uid := atoi(user["uid"])
	if uid == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	if _, err := s.store.CountVODSupports(ctx, uid); err != nil {
		return nil, -1, "获取助力求片记录失败", err
	}
	rows, err := s.store.VODSupports(ctx, uid, page, pageSize)
	if err != nil {
		return nil, -1, "获取助力求片记录失败", err
	}
	return map[string]interface{}{"data": rows}, 0, "", nil
}

func (s *Service) VODOrderHistoryOrders(ctx context.Context, token string, page int) (map[string]interface{}, int, string, error) {
	user, _, err := s.authenticatedUser(ctx, token)
	if err != nil {
		return nil, -9999, "您还没有登录", err
	}
	if atoi(user["uid"]) == 0 {
		return nil, -9999, "您还没有登录", nil
	}
	const pageSize = 20
	status := 1
	if _, err := s.store.CountVODOrders(ctx, 0, &status); err != nil {
		return nil, -1, "获取历史求片失败", err
	}
	rows, err := s.store.VODOrders(ctx, 0, &status, page, pageSize, "")
	if err != nil {
		return nil, -1, "获取历史求片失败", err
	}
	return map[string]interface{}{"data": rows}, 0, "", nil
}
