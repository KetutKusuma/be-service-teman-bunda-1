package controllers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/tensuqiuwulu/be-service-teman-bunda/config"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/request"
	"github.com/tensuqiuwulu/be-service-teman-bunda/models/http/response"
	"github.com/tensuqiuwulu/be-service-teman-bunda/services"
)

type AuthControllerInterface interface {
	Login(c echo.Context) error
	NewToken(c echo.Context) error
}

type AuthControllerImplementation struct {
	ConfigurationWebserver config.Webserver
	Logger                 *logrus.Logger
	AuthServiceInterface   services.AuthServiceInterface
}

func NewAuthController(configurationWebserver config.Webserver,
	logger *logrus.Logger,
	authServiceInterface services.AuthServiceInterface) AuthControllerInterface {
	return &AuthControllerImplementation{
		ConfigurationWebserver: configurationWebserver,
		Logger:                 logger,
		AuthServiceInterface:   authServiceInterface,
	}
}

func (controller *AuthControllerImplementation) Login(c echo.Context) error {
	requestId := ""
	request := request.ReadFromAuthRequestBody(c, requestId, controller.Logger)
	loginResponse := controller.AuthServiceInterface.Login(requestId, request)
	respon := response.Response{Code: 200, Mssg: "success", Data: loginResponse, Error: []string{}}
	return c.JSON(http.StatusOK, respon)
}

func (controller *AuthControllerImplementation) NewToken(c echo.Context) error {
	requestId := ""
	refreshToken := c.FormValue("refresh_token")
	token := controller.AuthServiceInterface.NewToken(requestId, refreshToken)
	respon := response.Response{Code: 200, Mssg: "success", Data: token, Error: []string{}}
	return c.JSON(http.StatusOK, respon)
}