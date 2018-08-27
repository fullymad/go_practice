package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// TODO: Improvement - Add computed fields (like taxable_amount) into
// structures so that they can be filled out and a separate function
// can be called at the end to send output as text, json, etc.
type OptionGrant struct {
	Shares      uint    // No. of shares
	StrikePrice float32 // Cost to employee to buy a share
}
type RsuGrant struct {
	Shares uint    // No. of shares
	Price  float32 // Price at time of vesting (relevant when vested)
	//purchaseDate - TODO: to separate long/short term gain
	Vested bool // Used to separate taxable/non-taxable gain
}
type StockGrant struct {
	Symbol string
	Option []OptionGrant
	Rsu    []RsuGrant
}
type StockPrice struct {
	Symbol string
	Price  float32 // To store current price
}

// Reads input stock grant data from given json file into given slice
func GetStockGrant(input_file string, stock_grant *[]StockGrant) error {
	input_data, err := ioutil.ReadFile(input_file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(input_data, stock_grant)
	/*
		pretty_format, err := json.MarshalIndent(stock_grant, "", "  ")
		if err != nil {
			fmt.Println("Error pretty printing json data:", err)
		} else {
			fmt.Printf("%s\n", pretty_format)
		}*/

	return err
} // end GetStockGrant

// Returns the current price of given symbols
func GetCurrentPrice(symbols []string) ([]StockPrice, error) {
	var http_str strings.Builder
	var prices []StockPrice
	get_end_point :=
		"https://api.iextrading.com/1.0/tops/last?symbols="

	http_str.WriteString(get_end_point)
	http_str.WriteString(strings.Join(symbols, ","))

	res, err := http.Get(http_str.String())
	if err != nil {
		return prices, err
	}

	if res.StatusCode != http.StatusOK {
		log.Fatalf("Error '%s' from GET on endpoint %s\n",
			http.StatusText(res.StatusCode), get_end_point)
	}

	stock_price, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return prices, err
	}

	err = json.Unmarshal(stock_price, &prices)
	if err != nil {
		fmt.Println("Error unmarshaling json data returned by endpoint")
	}

	return prices, err
} // end GetCurrentPrice

// For given company stock grant for one or more symbols, computes
// taxable and non-taxable amount based on current price
// ASSUMPTIONS:
//   All options are considered vested but not purchased
//   RSUs can be vested/unvested
//   ESPP not supported yet
func main() {
	var stock_grant []StockGrant
	var taxable float32 = 0.0
	var non_taxable float32 = 0.0
	var total_taxable float32 = 0.0
	var total_non_taxable float32 = 0.0
	var total_unvested float32 = 0.0
	var symbol_taxable float32 = 0.0
	var symbol_non_taxable float32 = 0.0
	var symbol_unvested float32 = 0.0
	var grand_total_taxable float32 = 0.0
	var grand_total_non_taxable float32 = 0.0
	var grand_total_unvested float32 = 0.0

	if len(os.Args) == 1 {
		log.Fatalln("Please provide stock grant file name as input")
	}
	stock_grant_file := os.Args[1]

	err := GetStockGrant(stock_grant_file, &stock_grant)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Optimize to get prices for all symbols in one call up front
	for _, sg := range stock_grant {
		symbol_taxable = 0.0
		symbol_non_taxable = 0.0
		symbol_unvested = 0.0
		current_price, err := GetCurrentPrice([]string{sg.Symbol})
		if err != nil {
			log.Fatal(err)
		}
		if len(current_price) == 0 {
			fmt.Printf("\nSymbol: %s - ERROR getting price, skipping\n",
				sg.Symbol)
			continue
		}
		fmt.Printf("\nSymbol: %s, Current price: %f\n", sg.Symbol,
			current_price[0].Price)

		// Process all options
		fmt.Printf("STOCK OPTIONS\n")
		total_taxable = 0.0
		for _, opt := range sg.Option {
			// Compute gain if exercised and sold at current price
			taxable = (float32)(opt.Shares) *
				(current_price[0].Price - opt.StrikePrice)

			// If under water, no gain or loss as no need to exercise
			if taxable < 0.0 {
				taxable = 0.0
			}
			fmt.Printf(
				"  Shares: %d\tStrike price: %.2f\tTaxable %.2f\n",
				opt.Shares, opt.StrikePrice, taxable)
			total_taxable += taxable
		} // end for option grants

		if sg.Option == nil { // No options
			fmt.Printf("  NONE\n")
		} else {
			symbol_taxable += total_taxable
			fmt.Printf("Total taxable %.2f\n", total_taxable)
		}

		// Process all RSUs
		fmt.Printf("RSU\n")
		total_taxable = 0.0
		total_non_taxable = 0.0
		total_unvested = 0.0
		for _, rsu := range sg.Rsu {
			if rsu.Vested {
				// Compute gain if sold at current price

				// Price at vesting already taxed
				non_taxable = (float32)(rsu.Shares) * rsu.Price

				// Difference with price at vesting will be taxed
				taxable = (float32)(rsu.Shares) *
					(current_price[0].Price - rsu.Price)
				total_taxable += taxable
				total_non_taxable += non_taxable
			} else {
				// Compute gain if sold (on vesting) at current price
				// Price at vesting will be taxed
				taxable = (float32)(rsu.Shares) * current_price[0].Price
				total_unvested += taxable
			}

			fmt.Printf("  Shares: %d\tVested: %t\tTaxable "+
				"%.2f\tNon-taxable %.2f\n",
				rsu.Shares, rsu.Vested, taxable, non_taxable)
		} // end for RSU grants

		if sg.Rsu == nil { // No RSU
			fmt.Printf("  NONE\n")
		} else {
			symbol_taxable += total_taxable
			symbol_non_taxable += total_non_taxable
			symbol_unvested += total_unvested
			fmt.Printf("Total taxable %.2f\t"+
				"Total non-taxable %.2f\n",
				total_taxable, total_non_taxable)
		}

		// Print totals for this symbol
		fmt.Printf("\n** %s Taxable %.2f\tNon-taxable "+
			"%.2f\tUnvested taxable %.2f\n", sg.Symbol, symbol_taxable,
			symbol_non_taxable, symbol_unvested)
		grand_total_taxable += symbol_taxable
		grand_total_non_taxable += symbol_non_taxable
		grand_total_unvested += symbol_unvested

	} // end for all symbols

	// Print grand totals for all symbols
	fmt.Printf("\n*** GRAND Totals: Taxable %.2f\t"+
		"Non-taxable %.2f\tUnvested taxable %.2f\n",
		grand_total_taxable, grand_total_non_taxable,
		grand_total_unvested)

} // end main
