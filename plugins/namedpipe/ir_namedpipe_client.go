package namedpipe

import (
	"fmt"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/namedpipe/namedpipecodegen"
)

// IRNode representing a client to a Golang NamedPipe server.
// This node does not introduce any new runtime interfaces or types that can be used by other IRNodes.
type GolangNamedPipeClient struct {
	golang.Node
	golang.Service
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName  string
	ServerAddr    *address.Address[*GolangNamedPipeServer]
	outputPackage string
}

func newGolangNamedPipeClient(name string, addr *address.Address[*GolangNamedPipeServer]) (*GolangNamedPipeClient, error) {
	node := &GolangNamedPipeClient{}
	node.InstanceName = name
	node.ServerAddr = addr
	node.outputPackage = "namedpipe"

	return node, nil
}

func (n *GolangNamedPipeClient) String() string {
	return n.InstanceName + " = NamedPipeClient(" + n.ServerAddr.Dial.Name() + ")"
}

func (n *GolangNamedPipeClient) Name() string {
	return n.InstanceName
}

func (node *GolangNamedPipeClient) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.ServerAddr.Server.GetInterface(ctx)
	if err != nil {
		return nil, err
	}

	np, isNp := iface.(*NamedPipeInterface)
	if !isNp {
		return nil, fmt.Errorf("namedpipe client expected a NamedPipe interface from %v but found %v", node.ServerAddr.Name(), iface)
	}

	wrapped, isValid := np.Wrapped.(*gocode.ServiceInterface)
	if !isValid {
		return nil, fmt.Errorf("namedpipe client expected the server's NamedPipe interface to wrap a gocode interface but found %v", np)
	}

	return wrapped, nil
}

// Just makes sure that the interface exposed by the server is included in the built module
func (node *GolangNamedPipeClient) AddInterfaces(builder golang.ModuleBuilder) error {
	return node.ServerAddr.Server.Wrapped.AddInterfaces(builder)
}

func (node *GolangNamedPipeClient) GenerateFuncs(builder golang.ModuleBuilder) error {
	if builder.Visited(node.InstanceName + ".generateFuncs") {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	return namedpipecodegen.GenerateClient(builder, iface, node.outputPackage)
}

func (node *GolangNamedPipeClient) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node)
	if err != nil {
		return err
	}

	// Here is the instantiation of the client...
	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_NamedPipeClient", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "addr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}

	// By passing an IRConfig in the parameter we basically tell it to generate the
	// IRClient and pass it to the constructor
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.ServerAddr.Dial})
}

func (node *GolangNamedPipeClient) ImplementsGolangNode()    {}
func (node *GolangNamedPipeClient) ImplementsGolangService() {}
