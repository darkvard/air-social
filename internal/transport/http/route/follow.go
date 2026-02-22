package route

// import (
//
//

// 	"github.com/gin-gonic/gin"

// 	"air-social/internal/transport/http/handler"
// 	"air-social/internal/transport/http/middleware"
// )

// func FollowRoutes(g *gin.RouterGroup, h *handler.FollowHandler, m *middleware.Manager) {
// 	group := g.Group(UserGroup).Group("", m.Auth)
// 	{
// 		group.POST(FollowUser, h.Follow)
// 		group.DELETE(FollowUser, h.Unfollow)
// 		group.GET(Followers, h.GetFollowers)
// 		group.GET(Followings, h.GetFollowings)
// 	}
// }
