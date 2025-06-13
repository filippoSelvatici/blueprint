package workloadgen

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"time"

	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
	"github.com/blueprint-uservices/blueprint/runtime/core/workload"
)

// Workload specific flags
var outfile = flag.String("outfile", "stats.csv", "Outfile where individual request information will be stored")
var duration = flag.String("duration", "5m", "Duration for which the workload should be run")
var tput = flag.Int64("tput", 64, "Desired throughput")

type ComplexWorkload interface {
	ImplementsComplexWorkload(ctx context.Context) error
}

type complexWldGen struct {
	ComplexWorkload

	frontend  frontend.Frontend
	items     []catalogue.Sock
	username  string
	password  string
	userId    string
	cardId    string
	addressId string
}

func NewComplexWorkload(ctx context.Context, frontend frontend.Frontend) (ComplexWorkload, error) {
	w := &complexWldGen{frontend: frontend}
	return w, nil
}

type FnType func() error

func statWrapper(apiName string, fn FnType) workload.Stat {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	s := workload.Stat{}
	s.Start = start.UnixNano()
	s.Duration = duration.Nanoseconds()
	s.IsError = (err != nil)
	s.ApiName = apiName
	return s
}

func (w *complexWldGen) RunGetCart(ctx context.Context) workload.Stat {
	sessionId := RandomSessionID()
	return statWrapper("GetCart", func() error {
		_, err := w.frontend.GetCart(ctx, sessionId)
		if err != nil {
			fmt.Printf("Error in GetCart: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunAddItem(ctx context.Context) workload.Stat {
	sessionId := RandomSessionID()
	itemID := w.items[0].ID
	return statWrapper("AddItem", func() error {
		_, err := w.frontend.AddItem(ctx, sessionId, itemID)
		if err != nil {
			fmt.Printf("Error in AddItem: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunListItems(ctx context.Context) workload.Stat {
	return statWrapper("ListItems", func() error {
		_, err := w.frontend.ListItems(ctx, []string{}, "", 1, 100)
		if err != nil {
			fmt.Printf("Error in ListItems: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunListTags(ctx context.Context) workload.Stat {
	return statWrapper("ListTags", func() error {
		_, err := w.frontend.ListTags(ctx)
		if err != nil {
			fmt.Printf("Error in ListTags: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunRegister(ctx context.Context) workload.Stat {
	first_name, last_name := GenRunRegister()
	return statWrapper("Register", func() error {
		_, err := w.frontend.Register(ctx, "", first_name+last_name, "password", first_name+"@blueprint.com", first_name, last_name)
		if err != nil {
			fmt.Printf("Error in Register: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunPostCard(ctx context.Context) workload.Stat {
	card := GenPostCard()
	return statWrapper("PostCard", func() error {
		_, err := w.frontend.PostCard(ctx, w.userId, card)
		if err != nil {
			fmt.Printf("Error in PostCard: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunPostAddress(ctx context.Context) workload.Stat {
	address := GenPostAddress()
	return statWrapper("PostAddress", func() error {
		_, err := w.frontend.PostAddress(ctx, w.userId, address)
		if err != nil {
			fmt.Printf("Error in PostAddress: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunAddItemToSpecificUser(ctx context.Context) workload.Stat {
	itemID := w.items[0].ID
	return statWrapper("AddItem", func() error {
		_, err := w.frontend.AddItem(ctx, w.userId, itemID)
		if err != nil {
			fmt.Printf("Error in AddItemToSpecificUser: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunNewOrder(ctx context.Context) workload.Stat {
	return statWrapper("NewOrder", func() error {
		// Using the userId as cartId
		_, err := w.frontend.NewOrder(ctx, w.userId, w.addressId, w.cardId, w.userId)
		if err != nil {
			fmt.Printf("Error in NewOrder: %v\n", err)
		}

		return err
	})
}

func (w *complexWldGen) RunAddItemAndNewOrder(ctx context.Context) workload.Stat {
	itemID := w.items[0].ID
	return statWrapper("AddItemAndNewOrder", func() error {
		// Using the userId as cartId
		// fmt.Println("About to send NewOrder request")
		_, err1 := w.frontend.AddItem(ctx, w.userId, itemID)
		_, err2 := w.frontend.NewOrder(ctx, w.userId, w.addressId, w.cardId, w.userId)
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
		// fmt.Println("Received response for NewOrder request")
		return nil
	})
}

func (w *complexWldGen) InitApp(ctx context.Context) {
	var err error
	for {
		if _, err = w.frontend.LoadCatalogue(ctx); err == nil {
			break
		}
		fmt.Println("Failed to load catalogue, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	for {
		if w.items, err = w.frontend.ListItems(ctx, []string{}, "", 1, 100); err == nil {
			break
		}
		fmt.Println("Failed to list the items, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	sort.Slice(w.items, func(i, j int) bool {
		return w.items[i].Quantity > w.items[j].Quantity
	})

	w.username = "harry"
	w.password = "password"
	for {
		if w.userId, err = w.frontend.Register(ctx, "", w.username, w.password, "harry@blueprint.com", "Harry", "Potter"); err == nil {
			break
		}
		fmt.Println("Failed to register user, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	// Add a card and an address for the user
	for {
		if w.cardId, err = w.frontend.PostCard(ctx, w.userId, GenPostCard()); err == nil {
			break
		}
		fmt.Println("Failed to post card, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}

	for {
		if w.addressId, err = w.frontend.PostAddress(ctx, w.userId, GenPostAddress()); err == nil {
			break
		}
		fmt.Println("Failed to post address, retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
}

func (w *complexWldGen) Run(ctx context.Context) error {
	w.InitApp(ctx)
	fmt.Println("Application initialized successfully.")

	wrk := workload.NewWorkload()

	// Configure the workload with the client side generators for the various APIs and their respective proportions
	wrk.AddAPI("GetCart", w.RunGetCart, 0)
	wrk.AddAPI("AddItem", w.RunAddItem, 0)
	wrk.AddAPI("ListItems", w.RunListItems, 100)
	wrk.AddAPI("ListTags", w.RunListTags, 0)
	wrk.AddAPI("Register", w.RunRegister, 0)
	wrk.AddAPI("PostCard", w.RunPostCard, 0)
	wrk.AddAPI("PostAddress", w.RunPostAddress, 0)
	wrk.AddAPI("AddItemToSpecificUser", w.RunAddItemToSpecificUser, 0)
	wrk.AddAPI("NewOrder", w.RunNewOrder, 0)
	wrk.AddAPI("AddItemAndNewOrder", w.RunAddItemAndNewOrder, 0)

	// Initialize the engine
	engine, err := workload.NewEngine(*outfile, *tput, *duration, wrk)
	if err != nil {
		return err
	}
	// Run the workload
	engine.RunOpenLoop(ctx)
	// Print statistics from the workload
	return engine.PrintStats()
}

func (w *complexWldGen) ImplementsComplexWorkload(context.Context) error {
	return nil
}
