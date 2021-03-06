package service

import (
	"context"
	"im/internal/user/cache"
	"im/internal/user/dao"
	"im/internal/user/model"
	"im/pkg/gerrors"
	"im/pkg/pb"
	"im/pkg/rpc_cli"
	"time"
)

type authService struct{}

var AuthService = new(authService)

// SignIn 长连接登录
func (*authService) SignIn(ctx context.Context, phoneNumber, code string, deviceId int64) (int64, string, error) {
	if !Verify(phoneNumber, code) {
		return 0, "", gerrors.ErrBadCode
	}

	user, err := dao.UserDao.GetByPhoneNumber(phoneNumber)
	if err != nil {
		return 0, "", err
	}

	if user == nil {
		user = &model.User{PhoneNumber: phoneNumber}
		id, err := dao.UserDao.Add(*user)
		if err != nil {
			return 0, "", err
		}
		user.Id = id
	}

	resp, err := rpc_cli.LogicIntClient.GetDevice(ctx, &pb.GetDeviceReq{DeviceId: deviceId})
	if err != nil {
		return 0, "", err
	}

	// 方便测试
	token := "0"
	//token := util.RandString(40)
	err = cache.AuthCache.Set(user.Id, resp.Device.DeviceId, model.Device{
		Type:   resp.Device.Type,
		Token:  token,
		Expire: time.Now().AddDate(0, 3, 0).Unix(),
	})
	if err != nil {
		return 0, "", err
	}

	return user.Id, token, nil
}

func Verify(phoneNumber, code string) bool {
	// 假装他成功了
	return true
}

// Auth 验证用户是否登录
func (*authService) Auth(ctx context.Context, userId, deviceId int64, token string) error {
	device, err := cache.AuthCache.Get(userId, deviceId)
	if err != nil {
		return err
	}

	if device == nil {
		return gerrors.ErrUnauthorized
	}

	if device.Expire < time.Now().Unix() {
		return gerrors.ErrUnauthorized
	}
	return nil
}
