package service

import "fmt"

type ExampleService struct{}

func (s *ExampleService) AddExample(userID uint) string {
	return fmt.Sprintf("JWT 校验通过，user_id=%d（示例接口，可改为真实逻辑）", userID)
}
