package specs

import (
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/wiring"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/cmplx_workload/workloadgen"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/cart"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/catalogue"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/frontend"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/order"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/payment"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/queuemaster"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/shipping"
	"github.com/blueprint-uservices/blueprint/examples/sockshop/workflow/user"

	"github.com/blueprint-uservices/blueprint/plugins/clientpool"
	"github.com/blueprint-uservices/blueprint/plugins/cmdbuilder"
	"github.com/blueprint-uservices/blueprint/plugins/goproc"
	"github.com/blueprint-uservices/blueprint/plugins/gotests"
	"github.com/blueprint-uservices/blueprint/plugins/http"
	"github.com/blueprint-uservices/blueprint/plugins/linuxcontainer"
	"github.com/blueprint-uservices/blueprint/plugins/mysql"
	"github.com/blueprint-uservices/blueprint/plugins/namedpipe"

	"github.com/blueprint-uservices/blueprint/plugins/simple"
	"github.com/blueprint-uservices/blueprint/plugins/workflow"
	"github.com/blueprint-uservices/blueprint/plugins/workload"

	"github.com/blueprint-uservices/blueprint/plugins/mongodb"
)

// A wiring spec that deploys each service to a separate process, with services communicating over named pipes.
// The user, cart, shipping, and order services use NoSQL databases deployed in separate containers.
// The catalogue service uses a sqlite database deployed in a separate container.
// The shipping service and queue master service run within the same process (TODO: separate processes)
var NamedPipes = cmdbuilder.SpecOption{
	Name:        "namedpipes",
	Description: "Deploys each service in a separate process with named pipes for communication.",
	Build:       makeNamedPipesSpec,
}

func makeNamedPipesSpec(spec wiring.WiringSpec) ([]string, error) {

	// Modifiers that will be applied to all services
	applyDefaults := func(serviceName string, useHttp ...bool) {
		// Golang-level modifiers that add functionality
		clientpool.Create(spec, serviceName, 10)
		if len(useHttp) > 0 && useHttp[0] {
			http.Deploy(spec, serviceName)
		} else {
			namedpipe.Deploy(spec, serviceName)
		}

		// Deploying to namespaces
		goproc.Deploy(spec, serviceName)

		// Also add to tests
		gotests.Test(spec, serviceName)
	}

	user_db := mongodb.Container(spec, "user_db")
	user_service := workflow.Service[user.UserService](spec, "user_service", user_db)
	applyDefaults(user_service)

	payment_service := workflow.Service[payment.PaymentService](spec, "payment_service", "5000000")
	applyDefaults(payment_service)

	cart_db := mongodb.Container(spec, "cart_db")
	cart_service := workflow.Service[cart.CartService](spec, "cart_service", cart_db)
	applyDefaults(cart_service)

	shipqueue := simple.Queue(spec, "shipping_queue")
	shipdb := mongodb.Container(spec, "shipping_db")
	shipping_service := workflow.Service[shipping.ShippingService](spec, "shipping_service", shipqueue, shipdb)
	applyDefaults(shipping_service)

	// Deploy queue master to the same process as the shipping proc
	// TODO: after distributed queue is supported, move to separate containers
	queue_master := workflow.Service[queuemaster.QueueMaster](spec, "queue_master", shipqueue, shipping_service)
	goproc.AddToProcess(spec, "shipping_proc", queue_master)

	order_db := mongodb.Container(spec, "order_db")
	order_service := workflow.Service[order.OrderService](spec, "order_service", user_service, cart_service, payment_service, shipping_service, order_db)
	applyDefaults(order_service)

	catalogue_db := mysql.Container(spec, "catalogue_db")
	catalogue_service := workflow.Service[catalogue.CatalogueService](spec, "catalogue_service", catalogue_db)
	applyDefaults(catalogue_service)

	frontend_service := workflow.Service[frontend.Frontend](spec, "frontend", user_service, catalogue_service, cart_service, order_service)
	applyDefaults(frontend_service, true) // Only the frontend gets deployed with HTTP

	// TODO blueprint does not seem to allow directly deploying wlgen in a container. Therefore its deployment is separate from the
	// Docker compose deployment of the other services and cannot benefit from Docker compose's name resolution: a hack for now is
	// to hardcode localhost in the client boilerplate of the http plugin (which is used to send the requests to the HTTP server of
	// the frontend service)
	wlgen := workload.Generator[workloadgen.ComplexWorkload](spec, "wlgen", frontend_service)

	linuxcontainer.CreateContainer(spec, "all_services_ctr", user_service, payment_service, cart_service, shipping_service, order_service, catalogue_service, frontend_service)
	// Instantiate starting with the frontend which will trigger all other services to be instantiated
	// Also include the tests and wlgen
	return []string{"all_services_ctr", wlgen, "gotests"}, nil
}
