package route

// import (
//
 

// 	"github.com/gin-gonic/gin"

// 	"air-social/internal/transport/http/handler"
// 	"air-social/internal/transport/http/middleware"
// )

// func PostRoutes(r *gin.RouterGroup, h *handler.PostHandler, m *middleware.Manager) {
// 	post := r.Group(PostGroup, m.Auth)
// 	{
// 		post.GET(ID, h.GetPost)
// 		post.DELETE(ID, h.DeletePost)

// 		json := post.Group("").Use(m.JSONOnly)
// 		{
// 			json.PATCH(ID, h.UpdatePost)
// 			json.POST("", h.CreatePost)
// 		}
// 	}

// 	user := r.Group(UserGroup, m.Auth)
// 	{
// 		user.GET(ID+PostGroup, h.GetUserPosts)
// 	}
// }
