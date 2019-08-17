// +build stripe

package apppaymentintent_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/stripe/stripe-go/paymentintent"

	apppaymentintent "github.com/lelledaniele/upaygo/payment/intent"
	"github.com/stripe/stripe-go/customer"

	appamount "github.com/lelledaniele/upaygo/amount"
	appcurrency "github.com/lelledaniele/upaygo/currency"
	appcustomer "github.com/lelledaniele/upaygo/customer"
	apppaymentsource "github.com/lelledaniele/upaygo/payment/source"

	appconfig "github.com/lelledaniele/upaygo/config"
)

func TestMain(m *testing.M) {
	var fcp string

	flag.StringVar(&fcp, "config", "", "Provide config file as an absolute path")
	flag.Parse()

	if fcp == "" {
		fmt.Print("Integration Stripe test needs the config file absolute path as flag -config")
		os.Exit(1)
	}
	fmt.Printf("provided path was %s\n", fcp)

	fc, e := os.Open(fcp)
	if e != nil {
		fmt.Printf("Impossible to get configuration file: %v\n", e)
		os.Exit(1)
	}
	defer fc.Close()

	e = appconfig.ImportConfig(fc)
	if e != nil {
		fmt.Printf("Error durring file config import: %v", e)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	cur, _ := appcurrency.New("EUR")
	am := 2045
	cus, _ := appcustomer.NewStripe("email@email.com", cur)
	a, _ := appamount.New(am, cur.GetISO4217())
	ps := apppaymentsource.New("pm_card_visa")

	pi, e := apppaymentintent.New(a, ps, cus)
	if e != nil {
		t.Errorf("impossible to create a new payment intent: %v", e)
	}

	if pi.GetGatewayReference() == "" {
		t.Error("intent new is incorrect, created an intent without gateway reference")
	}

	if pi.GetConfirmationMethod() != "manual" {
		t.Errorf("intent should have confirmation method set to manual, got %v", pi.GetConfirmationMethod())
	}

	if pi.GetNextAction() != "" {
		t.Errorf("intent should not have next action as it is not confirmed, got %v", pi.GetNextAction())
	}

	if !pi.IsOffSession() {
		t.Error("intent should be enable for off session payment")
	}

	if pi.GetCreatedAt().Unix() == 0 {
		t.Error("intent should have a create timestamp")
	}

	if pi.GetCustomer().GetGatewayReference() != cus.GetGatewayReference() {
		t.Errorf("intent customer is incorrect, got: %v want %v", pi.GetCustomer(), cus)
	}

	if pi.GetSource().GetGatewayReference() != ps.GetGatewayReference() {
		t.Errorf("intent source is incorrect, got: %v want %v", pi.GetSource(), ps)
	}

	if pi.GetAmount() != a {
		t.Errorf("intent amount is incorrect, got: %v want %v", pi.GetAmount(), a)
	}

	if pi.IsCanceled() {
		t.Error("a new intent should not be cancelled")
	}

	if pi.IsSucceeded() {
		t.Error("a new intent should not be succeeded")
	}

	if pi.RequiresCapture() {
		t.Error("a new intent should not require capture")
	}

	if !pi.RequiresConfirmation() {
		t.Error("a new intent should require confirmation")
	}

	_, _ = paymentintent.Cancel(pi.GetGatewayReference(), nil)
	_, _ = customer.Del(cus.GetGatewayReference(), nil)
}

func TestNewWithoutCustomer(t *testing.T) {
	cur, _ := appcurrency.New("EUR")
	am := 2045
	a, _ := appamount.New(am, cur.GetISO4217())
	ps := apppaymentsource.New("pm_card_visa")

	pi, e := apppaymentintent.New(a, ps, nil)
	if e != nil {
		t.Errorf("impossible to create a new payment intent: %v", e)
	}

	if pi.GetCustomer() != nil {
		t.Errorf("intent customer should be blank, got: %v", pi.GetCustomer())
	}

	_, _ = paymentintent.Cancel(pi.GetGatewayReference(), nil)
}

func TestNewWithoutAmount(t *testing.T) {
	ps := apppaymentsource.New("pm_card_visa")

	_, e := apppaymentintent.New(nil, ps, nil)
	if e == nil {
		t.Error("intent without amount created")
	}
}

func TestNewWithoutPaymentSource(t *testing.T) {
	am := 2045
	a, _ := appamount.New(am, "EUR")

	_, e := apppaymentintent.New(a, nil, nil)
	if e == nil {
		t.Error("intent without amount created")
	}
}