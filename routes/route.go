package routes

import (
	"github.com/labstack/echo/v4"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/controllers"
	authMiddlerware "github.com/tensuqiuwulu/be-service-teman-bunda/middleware"
)

// Provinsi Route
func ProvinsiRoute(e *echo.Echo, configWebserver config.Webserver, configJWT config.Jwt, provinsiControllerInterface controllers.ProvinsiControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/provinsi", provinsiControllerInterface.FindAllProvinsi)
}

// Kabupaten Route
func KabupatenRoute(e *echo.Echo, configWebserver config.Webserver, kabupatenControllerInterface controllers.KabupatenControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kabupaten", kabupatenControllerInterface.FindAllKabupatenByIdProvinsi)
}

// Kabupaten Route
func KecamatanRoute(e *echo.Echo, configWebserver config.Webserver, kecamatanControllerInterface controllers.KecamatanControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kecamatan", kecamatanControllerInterface.FindAllKecamatanByIdKabupaten)
}

// Kelurahan Route
func KelurahanRoute(e *echo.Echo, configWebserver config.Webserver, kelurahanControllerInterface controllers.KelurahanControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/kelurahan", kelurahanControllerInterface.FindAllKelurahanByIdKecamatan)
}

// User Route
func UserRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, userControllerInterface controllers.UserControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/user/create", userControllerInterface.CreateUser)
	group.GET("/user/referal", userControllerInterface.FindUserByReferal)
	group.GET("/user", userControllerInterface.FindUserById, authMiddlerware.Authentication(configurationJWT))
	group.POST("/user/verifyemail", userControllerInterface.SendVerifyEmail)
}

// Auth Route
func AuthRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, authControllerInterface controllers.AuthControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/auth/login", authControllerInterface.Login)
	group.POST("/auth/new-token", authControllerInterface.NewToken)
}

// Balance Point Route
func BalancePointRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, balancePointControllerInterface controllers.BalancePointControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/balance_point", balancePointControllerInterface.FindBalancePointByIdUser, authMiddlerware.Authentication(configurationJWT))
	group.GET("/balance_point/check/amount", balancePointControllerInterface.BalancePointCheckAmount, authMiddlerware.Authentication(configurationJWT))
	group.GET("/balance_point/check/order_tx", balancePointControllerInterface.BalancePointCheckOrderTx, authMiddlerware.Authentication(configurationJWT))
}

// Balance Point Tx
func BalancePointTxRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, balancePointTxControllerInterface controllers.BalancePointTxControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/balance_point_tx", balancePointTxControllerInterface.FindBalancePointTxByIdBalancePoint, authMiddlerware.Authentication(configurationJWT))
}

// Product Route
func ProductRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, productControllerInterface controllers.ProductControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/products", productControllerInterface.FindAllProducts, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/search", productControllerInterface.FindProductsBySearch, authMiddlerware.Authentication(configurationJWT))
	group.GET("/product/id", productControllerInterface.FindProductById, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/category", productControllerInterface.FindProductByIdCategory, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/sub_category", productControllerInterface.FindProductByIdSubCategory, authMiddlerware.Authentication(configurationJWT))
	group.GET("/products/brand", productControllerInterface.FindProductByIdBrand, authMiddlerware.Authentication(configurationJWT))
}

// Cart Route
func CartRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, cartControllerInterface controllers.CartControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/cart", cartControllerInterface.FindCartByIdUser, authMiddlerware.Authentication(configurationJWT))
	group.POST("/cart", cartControllerInterface.AddProductToCart, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/plus_qty", cartControllerInterface.CartPlusQtyProduct, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/min_qty", cartControllerInterface.CartMinQtyProduct, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/cart/update_qty", cartControllerInterface.UpdateQtyProductInCart, authMiddlerware.Authentication(configurationJWT))
}

// Shipping Cost Route
func ShippingRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, shippingControllerInterface controllers.ShippingControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/shipping/cost", shippingControllerInterface.GetShippingCostByIdKelurahan, authMiddlerware.Authentication(configurationJWT))
}

// Order Route
func OrderRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, orderControllerInterface controllers.OrderControllerInterface) {
	group := e.Group("api/v1")
	group.POST("/order/create", orderControllerInterface.CreateOrder, authMiddlerware.Authentication(configurationJWT))
	group.POST("/order/update", orderControllerInterface.UpdateStatusOrder)
	group.GET("/order", orderControllerInterface.FindOrderByUser, authMiddlerware.Authentication(configurationJWT))
	group.GET("/order/detail/id", orderControllerInterface.FindOrderById, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/order/cancel/id", orderControllerInterface.CancelOrderById, authMiddlerware.Authentication(configurationJWT))
	group.PUT("/order/complete/id", orderControllerInterface.CompleteOrderById, authMiddlerware.Authentication(configurationJWT))
}

// List Payment
func PaymentChannelRoute(e *echo.Echo, configWebserver config.Webserver, configurationJWT config.Jwt, paymentChannelControllerInterface controllers.PaymentChannelControllerInterface) {
	group := e.Group("api/v1")
	group.GET("/payment/list", paymentChannelControllerInterface.FindListPaymentChannel, authMiddlerware.Authentication(configurationJWT))
}
