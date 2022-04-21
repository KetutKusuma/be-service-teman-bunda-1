package response

import (
	"encoding/json"
	"fmt"

	"github.com/tensuqiuwulu/be-service-teman-bunda/models/entity"
)

type FindCartByIdUserResponse struct {
	SubTotal     float64    `json:"sub_total"`
	ShippingCost float64    `json:"shipping_cost"`
	TotalBill    float64    `json:"total_bill"`
	CartItems    []CartItem `json:"cart_items"`
}

type CartItem struct {
	Id          string  `json:"id"`
	IdProduct   string  `json:"id_product"`
	Price       float64 `json:"price"`
	ProductName string  `json:"product_name"`
	Description string  `json:"description"`
	Stock       int     `json:"stock"`
	PictureUrl  string  `json:"picture_url"`
	Thumbnail   string  `json:"thumbnail"`
	Qty         int     `json:"qty"`
	FlagPromo   string  `json:"flag_promo"`
}

func ToFindCartByIdUserResponse(carts []entity.Cart) (cartResponse FindCartByIdUserResponse) {

	var cartItems []CartItem
	var test CartItem
	hasil, _ := json.Marshal(test)
	fmt.Println(string(hasil))
	var totalPricePerItem float64
	var SubTotal float64
	var shippingCost float64
	for _, cart := range carts {
		var cartItem CartItem
		if cart.Product.ProductDiscount.FlagPromo == "true" {
			totalPricePerItem = cart.Product.ProductDiscount.Nominal
		} else {
			totalPricePerItem = cart.Product.Price
		}

		cartItem.Id = cart.Id
		cartItem.IdProduct = cart.IdProduct
		totalPricePerItem = totalPricePerItem * (float64(cart.Qty))
		cartItem.Price = totalPricePerItem
		cartItem.ProductName = cart.Product.ProductName
		cartItem.Description = cart.Product.Description
		cartItem.Stock = cart.Product.Stock
		cartItem.PictureUrl = cart.Product.PictureUrl
		cartItem.Thumbnail = cart.Product.Thumbnail
		cartItem.Qty = cart.Qty
		cartItem.FlagPromo = cart.Product.ProductDiscount.FlagPromo
		SubTotal = SubTotal + totalPricePerItem

		cartItems = append(cartItems, cartItem)
	}

	cartResponse.CartItems = cartItems
	cartResponse.SubTotal = SubTotal
	cartResponse.TotalBill = SubTotal + shippingCost

	return cartResponse
}
