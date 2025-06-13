package namedpipe

import (
	"fmt"
	"reflect"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/address"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/coreplugins/service"
	"github.com/blueprint-uservices/blueprint/blueprint/pkg/ir"
	"github.com/blueprint-uservices/blueprint/plugins/golang"
	"github.com/blueprint-uservices/blueprint/plugins/golang/gocode"
	"github.com/blueprint-uservices/blueprint/plugins/namedpipe/namedpipecodegen"
)

// IRNode representing a Golang NamedPipe server.
// This node does not introduce any new runtime interfaces or types that can be used by other IRNodes.
type GolangNamedPipeServer struct {
	service.ServiceNode
	golang.GeneratesFuncs
	golang.Instantiable

	InstanceName string
	Bind         *address.BindConfig
	Wrapped      golang.Service

	outputPackage string
}

// Represents a service that is exposed over named pipes.
type NamedPipeInterface struct {
	service.ServiceInterface
	Wrapped service.ServiceInterface
}

func (i *NamedPipeInterface) GetName() string {
	return "namedpipe(" + i.Wrapped.GetName() + ")"
}

func (i *NamedPipeInterface) GetMethods() []service.Method {
	return i.Wrapped.GetMethods()
}

func newGolangNamedPipeServer(name string, wrapped ir.IRNode) (*GolangNamedPipeServer, error) {
	service, is_service := wrapped.(golang.Service)
	if !is_service {
		return nil, blueprint.Errorf("Named pipe server %s expected %s to be a golang service, but got %s", name, wrapped.Name(), reflect.TypeOf(wrapped).String())
	}

	node := &GolangNamedPipeServer{}
	node.InstanceName = name
	node.Wrapped = service
	node.outputPackage = "namedpipe"
	return node, nil
}

func (n *GolangNamedPipeServer) String() string {
	return n.InstanceName + " = NamedPipeServer(" + n.Wrapped.Name() + ", " + n.Bind.Name() + ")"
}

func (n *GolangNamedPipeServer) Name() string {
	return n.InstanceName
}

// Generates the Named Pipe Server handler
func (node *GolangNamedPipeServer) GenerateFuncs(builder golang.ModuleBuilder) error {
	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	err = namedpipecodegen.GenerateServerHandler(builder, iface, node.outputPackage)
	if err != nil {
		return err
	}
	return nil
}

func (node *GolangNamedPipeServer) AddInstantiation(builder golang.NamespaceBuilder) error {
	// Only generate instantiation code for this instance once
	if builder.Visited(node.InstanceName) {
		return nil
	}

	iface, err := golang.GetGoInterface(builder, node.Wrapped)
	if err != nil {
		return err
	}

	constructor := &gocode.Constructor{
		Package: builder.Module().Info().Name + "/" + node.outputPackage,
		Func: gocode.Func{
			Name: fmt.Sprintf("New_%v_NamedPipeServerHandler", iface.BaseName),
			Arguments: []gocode.Variable{
				{Name: "ctx", Type: &gocode.UserType{Package: "context", Name: "Context"}},
				{Name: "service", Type: iface},
				{Name: "serverAddr", Type: &gocode.BasicType{Name: "string"}},
			},
		},
	}
	return builder.DeclareConstructor(node.InstanceName, constructor, []ir.IRNode{node.Wrapped, node.Bind})
}

func (node *GolangNamedPipeServer) GetInterface(ctx ir.BuildContext) (service.ServiceInterface, error) {
	iface, err := node.Wrapped.GetInterface(ctx)
	return &NamedPipeInterface{Wrapped: iface}, err
}

func (node *GolangNamedPipeServer) ImplementsGolangNode() {}
