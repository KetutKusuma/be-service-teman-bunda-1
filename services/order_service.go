package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/exceptions"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/request"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	modelService "github.com/tensuqiuwulu/be-service-teman-bunda/models/service"
	"github.com/tensuqiuwulu/be-service-teman-bunda/repository/mysql"
	"github.com/tensuqiuwulu/be-service-teman-bunda/utilities"
	"gorm.io/gorm"
)

type OrderServiceInterface interface {
	CreateOrder(requestId string, idUser string, orderRequest *request.CreateOrderRequest) (orderResponse response.CreateOrderResponse)
	UpdateStatusOrder(requestId string, orderRequest *request.CallBackIpaymuRequest) (orderResponse response.UpdateOrderStatusResponse)
	FindOrderByUser(requestId string, idUser string, orderStatus string) (orderResponses []response.FindOrderByUserResponse)
	FindOrderById(requestId string, idOrder string) (orderResponse response.FindOrderByNumberOrderResponse)
}

type OrderServiceImplementation struct {
	ConfigurationWebserver                 config.Webserver
	DB                                     *gorm.DB
	ConfigJwt                              config.Jwt
	Validate                               *validator.Validate
	Logger                                 *logrus.Logger
	ConfigurationIpaymu                    *config.Ipaymu
	OrderRepositoryInterface               mysql.OrderRepositoryInterface
	CartRepositoryInterface                mysql.CartRepositoryInterface
	UserRepositoryInterface                mysql.UserRepositoryInterface
	OrderItemRepositoryInterface           mysql.OrderItemRepositoryInterface
	PaymentLogRepositoryInterface          mysql.PaymentLogRepositoryInterface
	BankTransferRepositoryInterface        mysql.BankTransferRepositoryInterface
	BankVaRepositoryInterface              mysql.BankVaRepositoryInterface
	ProductRepositoryInterface             mysql.ProductRepositoryInterface
	ProductStockHistoryRepositoryInterface mysql.ProductStockHistoryRepositoryInterface
	BalancePointRepositoryInterface        mysql.BalancePointRepositoryInterface
	BalancePointTxRepositoryInterface      mysql.BalancePointTxRepositoryInterface
}

func NewOrderService(
	configurationWebserver config.Webserver,
	DB *gorm.DB,
	configJwt config.Jwt,
	validate *validator.Validate,
	logger *logrus.Logger,
	configIpaymu *config.Ipaymu,
	orderRepositoryInterface mysql.OrderRepositoryInterface,
	cartRepositoryInterface mysql.CartRepositoryInterface,
	userRepositoryInterface mysql.UserRepositoryInterface,
	orderItemRepositoryInterface mysql.OrderItemRepositoryInterface,
	paymentLogRepositoryInterface mysql.PaymentLogRepositoryInterface,
	bankTransferRepositoryInterface mysql.BankTransferRepositoryInterface,
	bankVaRepositoryInterface mysql.BankVaRepositoryInterface,
	productRepositoryInterface mysql.ProductRepositoryInterface,
	productStockHistoryRepositoryInterface mysql.ProductStockHistoryRepositoryInterface,
	balancePointRepositoryInterface mysql.BalancePointRepositoryInterface,
	balancePointTxRepositoryInterface mysql.BalancePointTxRepositoryInterface) OrderServiceInterface {
	return &OrderServiceImplementation{
		ConfigurationWebserver:                 configurationWebserver,
		DB:                                     DB,
		ConfigJwt:                              configJwt,
		Validate:                               validate,
		Logger:                                 logger,
		ConfigurationIpaymu:                    configIpaymu,
		OrderRepositoryInterface:               orderRepositoryInterface,
		CartRepositoryInterface:                cartRepositoryInterface,
		UserRepositoryInterface:                userRepositoryInterface,
		OrderItemRepositoryInterface:           orderItemRepositoryInterface,
		PaymentLogRepositoryInterface:          paymentLogRepositoryInterface,
		BankTransferRepositoryInterface:        bankTransferRepositoryInterface,
		BankVaRepositoryInterface:              bankVaRepositoryInterface,
		ProductRepositoryInterface:             productRepositoryInterface,
		ProductStockHistoryRepositoryInterface: productStockHistoryRepositoryInterface,
		BalancePointRepositoryInterface:        balancePointRepositoryInterface,
		BalancePointTxRepositoryInterface:      balancePointTxRepositoryInterface,
	}
}

func (service *OrderServiceImplementation) FindOrderByUser(requestId string, numberOrder string, orderStatus string) (orderResponses []response.FindOrderByUserResponse) {
	orders, err := service.OrderRepositoryInterface.FindOrderByUser(service.DB, numberOrder, orderStatus)
	exceptions.PanicIfError(err, requestId, service.Logger)
	orderResponses = response.ToFindOrderByUserResponse(orders)
	return orderResponses
}

func (service *OrderServiceImplementation) FindOrderById(requestId string, idOrder string) (orderResponse response.FindOrderByNumberOrderResponse) {
	order, err := service.OrderRepositoryInterface.FindOrderById(service.DB, idOrder)
	fmt.Println("order = ", order.ShippingCost)
	exceptions.PanicIfError(err, requestId, service.Logger)

	orderItems, err := service.OrderItemRepositoryInterface.FindOrderItemsByIdOrder(service.DB, idOrder)
	exceptions.PanicIfError(err, requestId, service.Logger)

	orderResponse = response.ToFindOrderByNumberOrder(order, orderItems)
	return orderResponse
}

func (service *OrderServiceImplementation) UpdateStatusOrder(requestId string, paymentRequestCallback *request.CallBackIpaymuRequest) (orderResponse response.UpdateOrderStatusResponse) {

	// validate request
	request.ValidateCallBackIpaymuRequest(service.Validate, paymentRequestCallback, requestId, service.Logger)

	order, _ := service.OrderRepositoryInterface.FindOrderByNumberOrder(service.DB, paymentRequestCallback.ReferenceId)

	if order.PaymentStatus == "Sudah Dibayar" {
		orderResponse = response.ToUpdateOrderStatusResponse(order)
		return orderResponse
	} else {
		if order.Id == "" {
			err := errors.New("order not found")
			exceptions.PanicIfRecordNotFound(err, requestId, []string{"order not found"}, service.Logger)
		}

		tx := service.DB.Begin()

		orderEntity := &entity.Order{}
		orderEntity.OrderSatus = "Menunggu Konfirmasi"
		if paymentRequestCallback.StatusCode == "1" {
			orderEntity.PaymentStatus = "Sudah Dibayar"
		} else {
			orderEntity.PaymentStatus = "Pending"
		}

		orderEntity.PaymentSuccessAt.Time = time.Now()

		orderResult, err := service.OrderRepositoryInterface.UpdateOrderStatus(tx, paymentRequestCallback.ReferenceId, *orderEntity)
		exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error update order"}, service.Logger, tx)

		//update product stock
		orderItems, _ := service.OrderItemRepositoryInterface.FindOrderItemsByIdOrder(service.DB, order.Id)
		for _, orderItem := range orderItems {
			productEntity := &entity.Product{}
			productEntityStockHistory := &entity.ProductStockHistory{}
			product, errFindProduct := service.ProductRepositoryInterface.FindProductById(tx, orderItem.IdProduct)
			exceptions.PanicIfErrorWithRollback(errFindProduct, requestId, []string{"product not found"}, service.Logger, tx)

			productEntityStockHistory.IdProduct = orderItem.IdProduct
			productEntityStockHistory.TxDate = time.Now()
			productEntityStockHistory.StockOpname = product.Stock
			productEntityStockHistory.StockOutQty = orderItem.Qty
			productEntityStockHistory.StockFinal = product.Stock - orderItem.Qty
			productEntityStockHistory.Description = "Pembelian " + order.NumberOrder
			productEntityStockHistory.CreatedAt = time.Now()
			_, errAddProductStockHistory := service.ProductStockHistoryRepositoryInterface.AddProductStockHistory(tx, *productEntityStockHistory)
			exceptions.PanicIfErrorWithRollback(errAddProductStockHistory, requestId, []string{"add stock history error"}, service.Logger, tx)

			productEntity.Stock = product.Stock - orderItem.Qty
			_, errUpdateProductStock := service.ProductRepositoryInterface.UpdateProductStock(tx, orderItem.IdProduct, *productEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateProductStock, requestId, []string{"update stock error"}, service.Logger, tx)
		}

		// Create Balance Point Tx
		if order.PaymentByPoint != 0 {
			// get data balance point
			balancePoint, _ := service.BalancePointRepositoryInterface.FindBalancePointByIdUser(service.DB, order.IdUser)

			// update balance point
			balancePointEntity := &entity.BalancePoint{}
			balancePointEntity.BalancePoints = balancePoint.BalancePoints - order.PaymentByPoint

			// add balance point tx history
			balancePointTxEntity := &entity.BalancePointTx{}
			balancePointTxEntity.Id = utilities.RandomUUID()
			balancePointTxEntity.IdBalancePoint = balancePoint.Id
			balancePointTxEntity.NoOrder = order.NumberOrder
			balancePointTxEntity.TxType = "Pembelian"
			balancePointTxEntity.TxDate = time.Now()
			balancePointTxEntity.TxNominal = order.PaymentByPoint - (order.PaymentByPoint * 2)
			balancePointTxEntity.LastPointBalance = balancePoint.BalancePoints
			balancePointTxEntity.NewPointBalance = balancePoint.BalancePoints - order.PaymentByPoint
			balancePointTxEntity.CreatedDate = time.Now()

			_, errUpdateBalancePoint := service.BalancePointRepositoryInterface.UpdateBalancePoint(tx, balancePoint.IdUser, *balancePointEntity)
			exceptions.PanicIfErrorWithRollback(errUpdateBalancePoint, requestId, []string{"update balance point error"}, service.Logger, tx)

			_, errCreateBalancePointTx := service.BalancePointTxRepositoryInterface.CreateBalancePointTx(tx, *balancePointTxEntity)
			exceptions.PanicIfErrorWithRollback(errCreateBalancePointTx, requestId, []string{"create balance point tx error"}, service.Logger, tx)
		}

		// Create response log
		paymentLogEntity := &entity.PaymentLog{}
		paymentLogEntity.Id = utilities.RandomUUID()
		paymentLogEntity.IdOrder = order.Id
		paymentLogEntity.NumberOrder = order.NumberOrder
		paymentLogEntity.TypeLog = "Respon Success Ipaymu"
		paymentLogEntity.PaymentMethod = order.PaymentMethod
		paymentLogEntity.PaymentChannel = order.PaymentChannel
		paymentLogEntity.Log = fmt.Sprintf("%+v\n", paymentRequestCallback)
		paymentLogEntity.CreatedAt = time.Now()

		// s := fmt.Sprintf("%+v\n", paymentRequestCallback)
		// fmt.Println(s)

		_, errCreateLog := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
		exceptions.PanicIfErrorWithRollback(errCreateLog, requestId, []string{"Error create log"}, service.Logger, tx)

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)
		orderResponse = response.ToUpdateOrderStatusResponse(orderResult)
		return orderResponse
	}
}

func (service *OrderServiceImplementation) GenerateNumberOrder() (numberOrder string) {
	now := time.Now()
	orderEntity := &entity.Order{}
	for {
		rand.Seed(time.Now().Unix())
		charSet := "0123456789"
		var output strings.Builder
		length := 7

		for i := 0; i < length; i++ {
			random := rand.Intn(len(charSet))
			randomChar := charSet[random]
			output.WriteString(string(randomChar))
		}

		orderEntity.NumberOrder = "ORDER/" + now.Format("20060102") + "/" + output.String()

		// Check referal code if exist
		checkNumberOrder, _ := service.OrderRepositoryInterface.FindOrderByNumberOrder(service.DB, orderEntity.NumberOrder)
		if checkNumberOrder.Id == "" {
			break
		}
	}
	return orderEntity.NumberOrder
}

func (service *OrderServiceImplementation) CreateOrder(requestId string, idUser string, orderRequest *request.CreateOrderRequest) (orderResponse response.CreateOrderResponse) {

	fmt.Println("waktu order = ", time.Now())
	// Validate request
	request.ValidateCreateOrderRequest(service.Validate, orderRequest, requestId, service.Logger)

	// Get data user
	user, _ := service.UserRepositoryInterface.FindUserById(service.DB, idUser)

	tx := service.DB.Begin()
	exceptions.PanicIfError(tx.Error, requestId, service.Logger)

	// Create Order
	orderEntity := &entity.Order{}
	orderEntity.Id = utilities.RandomUUID()
	orderEntity.IdUser = user.Id
	orderEntity.NumberOrder = service.GenerateNumberOrder()
	orderEntity.FullName = user.FamilyMembers.FullName
	orderEntity.Email = user.FamilyMembers.Email
	orderEntity.Address = orderRequest.Address
	orderEntity.Phone = user.FamilyMembers.Phone
	orderEntity.CourierNote = orderRequest.CourierNote
	orderEntity.TotalBill = orderRequest.TotalBill
	orderEntity.OrderSatus = "Menunggu Pembayaran"
	orderEntity.OrderedAt = time.Now()
	fmt.Println("waktu order 2 = ", orderEntity.OrderedAt)
	orderEntity.PaymentMethod = orderRequest.PaymentMethod
	orderEntity.PaymentChannel = orderRequest.PaymentChannel
	orderEntity.PaymentStatus = "Belum Dibayar"
	orderEntity.PaymentByPoint = orderRequest.PaymentByPoint
	orderEntity.PaymentByCash = orderRequest.PaymentByCash
	orderEntity.ShippingCost = orderRequest.ShippingCost
	orderEntity.ShippingStatus = "Menunggu"
	order, err := service.OrderRepositoryInterface.CreateOrder(tx, *orderEntity)
	exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create order"}, service.Logger, tx)

	// Get data cart
	cartItems, _ := service.CartRepositoryInterface.FindCartByIdUser(service.DB, idUser)

	// Create order items
	var totalPriceProduct float64
	var orderItems []entity.OrderItem
	for _, cartItem := range cartItems {
		orderItemEntity := &entity.OrderItem{}
		orderItemEntity.Id = utilities.RandomUUID()
		orderItemEntity.IdOrder = orderEntity.Id
		orderItemEntity.IdProduct = cartItem.IdProduct
		orderItemEntity.NoSku = cartItem.Product.NoSku
		orderItemEntity.ProductName = cartItem.Product.ProductName
		orderItemEntity.PictureUrl = cartItem.Product.PictureUrl
		orderItemEntity.Description = cartItem.Product.Description
		orderItemEntity.Weight = cartItem.Product.Weight
		orderItemEntity.Volume = cartItem.Product.Volume
		orderItemEntity.Qty = cartItem.Qty
		if cartItem.Product.ProductDiscount.FlagPromo == "true" {
			orderItemEntity.Price = cartItem.Product.ProductDiscount.Nominal
			totalPriceProduct = cartItem.Product.ProductDiscount.Nominal
		} else {
			orderItemEntity.Price = cartItem.Product.Price
			totalPriceProduct = cartItem.Product.Price
		}

		orderItemEntity.TotalPrice = totalPriceProduct * (float64(cartItem.Qty))
		orderItemEntity.CreatedAt = time.Now()
		orderItems = append(orderItems, *orderItemEntity)
	}

	errCreateOrderItem := service.OrderItemRepositoryInterface.CreateOrderItems(tx, orderItems)
	exceptions.PanicIfErrorWithRollback(errCreateOrderItem, requestId, []string{"Error create order"}, service.Logger, tx)

	// delete data item in cart
	errDelete := service.CartRepositoryInterface.DeleteAllProductInCartByIdUser(tx, idUser, cartItems)
	exceptions.PanicIfErrorWithRollback(errDelete, requestId, []string{"Error delete in cart"}, service.Logger, tx)

	// chose metode pembayaran
	switch orderRequest.PaymentMethod {
	case "va", "qris":
		// Send request to ipaymu
		var ipaymu_va = "0000007762212544"
		var ipaymu_key = "SANDBOXBA640645-B4FF-488B-A540-7F866791E73E-20220425110704"

		url, _ := url.Parse("https://sandbox.ipaymu.com/api/v2/payment/direct")

		postBody, _ := json.Marshal(map[string]interface{}{
			"name":           orderEntity.FullName,
			"phone":          orderEntity.Phone,
			"email":          orderEntity.Email,
			"amount":         orderEntity.PaymentByCash,
			"notifyUrl":      "http://117.53.44.216:9000/api/v1/order/update",
			"expired":        24,
			"expiredType":    "hours",
			"referenceId":    orderEntity.NumberOrder,
			"paymentMethod":  orderRequest.PaymentMethod,
			"paymentChannel": orderRequest.PaymentChannel,
		})

		bodyHash := sha256.Sum256([]byte(postBody))
		bodyHashToString := hex.EncodeToString(bodyHash[:])
		stringToSign := "POST:" + ipaymu_va + ":" + strings.ToLower(string(bodyHashToString)) + ":" + ipaymu_key

		h := hmac.New(sha256.New, []byte(ipaymu_key))
		h.Write([]byte(stringToSign))
		signature := hex.EncodeToString(h.Sum(nil))

		reqBody := ioutil.NopCloser(strings.NewReader(string(postBody)))

		req := &http.Request{
			Method: "POST",
			URL:    url,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
				"va":           {ipaymu_va},
				"signature":    {signature},
			},
			Body: reqBody,
		}

		resp, err := http.DefaultClient.Do(req)

		if err != nil {
			log.Fatalf("An Error Occured %v", err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}
		defer resp.Body.Close()

		// get data bank
		bankVa, _ := service.BankVaRepositoryInterface.FindBankVaByBankCode(service.DB, orderRequest.PaymentChannel)

		var dataResponseIpaymu modelService.PaymentResponse

		if err := json.NewDecoder(resp.Body).Decode(&dataResponseIpaymu); err != nil {
			fmt.Println(err)
			exceptions.PanicIfError(err, requestId, service.Logger)
		}

		if dataResponseIpaymu.Status != 200 {
			exceptions.PanicIfErrorWithRollback(errors.New("error response ipaymu"), requestId, []string{"Error response ipaymu"}, service.Logger, tx)
		} else if dataResponseIpaymu.Status == 200 {
			// make log
			paymentLogEntity := &entity.PaymentLog{}
			paymentLogEntity.Id = utilities.RandomUUID()
			paymentLogEntity.IdOrder = orderEntity.Id
			paymentLogEntity.NumberOrder = orderEntity.NumberOrder
			paymentLogEntity.TypeLog = "Create Trx Ipaymu"
			paymentLogEntity.PaymentMethod = orderRequest.PaymentMethod
			paymentLogEntity.PaymentChannel = orderRequest.PaymentChannel
			paymentLogEntity.Log = fmt.Sprintf("%+v\n", dataResponseIpaymu)
			paymentLogEntity.CreatedAt = time.Now()

			_, err := service.PaymentLogRepositoryInterface.CreatePaymentLog(tx, *paymentLogEntity)
			exceptions.PanicIfErrorWithRollback(err, requestId, []string{"Error create log"}, service.Logger, tx)

			commit := tx.Commit()
			exceptions.PanicIfError(commit.Error, requestId, service.Logger)
		}

		orderResponse = response.ToCreateOrderVaResponse(order, dataResponseIpaymu, bankVa)
		return orderResponse

	case "trf":
		// Get data bank by code
		bankTransfer, _ := service.BankTransferRepositoryInterface.FindBankTransferByBankCode(service.DB, orderRequest.PaymentChannel)
		if bankTransfer.Id == "" {
			exceptions.PanicIfErrorWithRollback(errors.New("bank not found"), requestId, []string{"Bank not found"}, service.Logger, tx)
		}

		payment := &modelService.PaymentResponse{}

		// buat 3 nomor acak
		rand.Seed(time.Now().UnixNano())
		min := 500
		max := 999
		randNumber := rand.Intn(max-min+1) + min

		payment.Data.Total = orderRequest.PaymentByCash + float64(randNumber)
		payment.Data.PaymentName = bankTransfer.BankName
		payment.Data.PaymentNo = bankTransfer.NoAccount
		payment.Data.ReferenceId = orderEntity.NumberOrder

		commit := tx.Commit()
		exceptions.PanicIfError(commit.Error, requestId, service.Logger)

		orderResponse = response.ToCreateOrderTransferResponse(order, *payment, bankTransfer)
		return orderResponse
	case "cod":
		orderResponse = response.ToCreateOrderCodResponse(order)
		return orderResponse
	default:
		exceptions.PanicIfErrorWithRollback(errors.New("payment method not found"), requestId, []string{"payment method not found"}, service.Logger, tx)
		return
	}
}
