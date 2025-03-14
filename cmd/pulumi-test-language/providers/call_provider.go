// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/blang/semver"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/rpcutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
	pulumirpc "github.com/pulumi/pulumi/sdk/v3/proto/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

type CallProvider struct {
	plugin.UnimplementedProvider
}

var _ plugin.Provider = (*CallProvider)(nil)

func (p *CallProvider) Close() error {
	return nil
}

func (p *CallProvider) SignalCancellation(context.Context) error {
	return nil
}

func (p *CallProvider) Pkg() tokens.Package {
	return "call"
}

func (p *CallProvider) GetPluginInfo(context.Context) (workspace.PluginInfo, error) {
	version := semver.MustParse("13.3.7")
	info := workspace.PluginInfo{Version: &version}
	return info, nil
}

func (p *CallProvider) GetSchema(context.Context, plugin.GetSchemaRequest) (plugin.GetSchemaResponse, error) {
	primitiveType := func(t string) schema.PropertySpec {
		return schema.PropertySpec{
			TypeSpec: schema.TypeSpec{
				Type: t,
			},
		}
	}

	refType := func(t string) schema.PropertySpec {
		return schema.PropertySpec{
			TypeSpec: schema.TypeSpec{
				Type: "ref",
				Ref:  t,
			},
		}
	}

	resource := func(isComponent bool) func(
		description string,
		inputs map[string]schema.PropertySpec,
		outputs map[string]schema.PropertySpec,
	) schema.ResourceSpec {
		return func(
			description string,
			inputs map[string]schema.PropertySpec,
			outputs map[string]schema.PropertySpec,
		) schema.ResourceSpec {
			requiredInputs := slices.Sorted(maps.Keys(inputs))
			requiredOutputs := slices.Sorted(maps.Keys(outputs))

			return schema.ResourceSpec{
				IsComponent: isComponent,
				ObjectTypeSpec: schema.ObjectTypeSpec{
					Description: description,
					Type:        "object",
					Properties:  outputs,
					Required:    requiredOutputs,
				},
				InputProperties: inputs,
				RequiredInputs:  requiredInputs,
			}
		}
	}

	customResource := resource(false)

	pkg := schema.PackageSpec{
		Name:      "call",
		Version:   "13.3.7",
		Functions: map[string]schema.FunctionSpec{},
		Resources: map[string]schema.ResourceSpec{},
		Provider: customResource(
			"A provider whose provider resource supports method calls",
			map[string]schema.PropertySpec{
				"value": primitiveType("string"),
			},
			map[string]schema.PropertySpec{
				"value": primitiveType("string"),
			},
		),
	}
	pkg.Functions["pulumi:providers:call/identity"] = schema.FunctionSpec{
		Description: "The `identity` method of the `call` package's provider. " +
			"Returns the provider's `value` configuration unaltered.",
		Inputs: &schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				"__self__": refType("#/resources/pulumi:providers:call"),
			},
			Required: []string{"__self__"},
		},
		ReturnType: &schema.ReturnTypeSpec{
			ObjectTypeSpec: &schema.ObjectTypeSpec{
				Type: "object",
				Properties: map[string]schema.PropertySpec{
					"result": primitiveType("string"),
				},
				Required: []string{"result"},
			},
		},
	}
	pkg.Functions["pulumi:providers:call/prefixed"] = schema.FunctionSpec{
		Description: "The `prefixed` method of the `call` package's provider. " +
			"Accepts a string and returns the provider's `value` configuration prefixed with that string.",
		Inputs: &schema.ObjectTypeSpec{
			Type: "object",
			Properties: map[string]schema.PropertySpec{
				"__self__": refType("#/resources/pulumi:providers:call"),
				"prefix":   primitiveType("string"),
			},
			Required: []string{"__self__", "prefix"},
		},
		ReturnType: &schema.ReturnTypeSpec{
			ObjectTypeSpec: &schema.ObjectTypeSpec{
				Type: "object",
				Properties: map[string]schema.PropertySpec{
					"result": primitiveType("string"),
				},
				Required: []string{"result"},
			},
		},
	}
	pkg.Provider.Methods = map[string]string{
		"identity": "pulumi:providers:call/identity",
		"prefixed": "pulumi:providers:call/prefixed",
	}

	jsonBytes, err := json.Marshal(pkg)
	if err != nil {
		return plugin.GetSchemaResponse{}, err
	}

	res := plugin.GetSchemaResponse{Schema: jsonBytes}
	return res, nil
}

func (p *CallProvider) GetMapping(
	context.Context, plugin.GetMappingRequest,
) (plugin.GetMappingResponse, error) {
	return plugin.GetMappingResponse{}, nil
}

func (p *CallProvider) GetMappings(
	context.Context, plugin.GetMappingsRequest,
) (plugin.GetMappingsResponse, error) {
	return plugin.GetMappingsResponse{}, nil
}

func (p *CallProvider) CheckConfig(
	_ context.Context,
	req plugin.CheckConfigRequest,
) (plugin.CheckConfigResponse, error) {
	version, ok := req.News["version"]
	if !ok {
		return plugin.CheckConfigResponse{
			Failures: makeCheckFailure("version", "missing version"),
		}, nil
	}

	if !version.IsString() {
		return plugin.CheckConfigResponse{
			Failures: makeCheckFailure("version", "version is not a string"),
		}, nil
	}

	if version.StringValue() != "13.3.7" {
		return plugin.CheckConfigResponse{
			Failures: makeCheckFailure("version", "version is not 13.3.7"),
		}, nil
	}

	// version and value
	if len(req.News) > 2 {
		return plugin.CheckConfigResponse{
			Failures: makeCheckFailure("", fmt.Sprintf("too many properties: %v", req.News)),
		}, nil
	}

	return plugin.CheckConfigResponse{Properties: req.News}, nil
}

func (p *CallProvider) DiffConfig(
	context.Context, plugin.DiffConfigRequest,
) (plugin.DiffConfigResponse, error) {
	return plugin.DiffResult{}, nil
}

func (p *CallProvider) Configure(context.Context, plugin.ConfigureRequest) (plugin.ConfigureResponse, error) {
	return plugin.ConfigureResponse{}, nil
}

func (p *CallProvider) Check(
	_ context.Context,
	req plugin.CheckRequest,
) (plugin.CheckResponse, error) {
	return plugin.CheckResponse{
		Failures: makeCheckFailure("", fmt.Sprintf("invalid URN type: %s", req.URN.Type())),
	}, nil
}

func (p *CallProvider) Diff(
	context.Context, plugin.DiffRequest,
) (plugin.DiffResponse, error) {
	return plugin.DiffResult{}, nil
}

func (p *CallProvider) Create(
	_ context.Context,
	req plugin.CreateRequest,
) (plugin.CreateResponse, error) {
	return plugin.CreateResponse{Status: resource.StatusUnknown}, fmt.Errorf("invalid URN type: %s", req.URN.Type())
}

func (p *CallProvider) Call(
	ctx context.Context,
	req plugin.CallRequest,
) (plugin.CallResponse, error) {
	conn, err := grpc.NewClient(
		req.Info.MonitorAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		rpcutil.GrpcChannelOptions(),
	)
	if err != nil {
		return plugin.CallResponse{}, fmt.Errorf("connect to resource monitor: %w", err)
	}
	defer conn.Close()

	monitor := pulumirpc.NewResourceMonitorClient(conn)
	if req.Tok == "pulumi:providers:call/identity" {
		return p.callProviderIdentity(ctx, req, monitor)
	} else if req.Tok == "pulumi:providers:call/prefixed" {
		return p.callProviderPrefixed(ctx, req, monitor)
	}

	return plugin.CallResponse{}, fmt.Errorf("unknown function %v", req.Tok)
}

func (p *CallProvider) callProviderIdentity(
	ctx context.Context,
	req plugin.CallRequest,
	monitor pulumirpc.ResourceMonitorClient,
) (plugin.CallResponse, error) {
	selfRef := req.Args["__self__"].ResourceReferenceValue()

	selfRes, err := monitor.Invoke(ctx, &pulumirpc.ResourceInvokeRequest{
		Tok: "pulumi:pulumi:getResource",
		Args: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"urn": structpb.NewStringValue(string(selfRef.URN)),
			},
		},
		AcceptResources: true,
	})
	if err != nil {
		return plugin.CallResponse{}, fmt.Errorf("hydrating __self__ resource reference: %w", err)
	}

	value := selfRes.Return.Fields["state"].GetStructValue().Fields["value"]
	result := value.GetStringValue()

	return plugin.CallResponse{
		Return: resource.NewPropertyMapFromMap(map[string]interface{}{
			"result": result,
		}),
	}, nil
}

func (p *CallProvider) callProviderPrefixed(
	ctx context.Context,
	req plugin.CallRequest,
	monitor pulumirpc.ResourceMonitorClient,
) (plugin.CallResponse, error) {
	prefix, ok := req.Args["prefix"]
	if !ok {
		return plugin.CallResponse{
			Failures: makeCheckFailure("prefix", "missing prefix"),
		}, nil
	}

	if !prefix.IsString() {
		return plugin.CallResponse{
			Failures: makeCheckFailure("prefix", "prefix is not a string"),
		}, nil
	}

	selfRef := req.Args["__self__"].ResourceReferenceValue()

	selfRes, err := monitor.Invoke(ctx, &pulumirpc.ResourceInvokeRequest{
		Tok: "pulumi:pulumi:getResource",
		Args: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"urn": structpb.NewStringValue(string(selfRef.URN)),
			},
		},
		AcceptResources: true,
	})
	if err != nil {
		return plugin.CallResponse{}, fmt.Errorf("hydrating __self__ resource reference: %w", err)
	}

	value := selfRes.Return.Fields["state"].GetStructValue().Fields["value"]
	result := prefix.StringValue() + value.GetStringValue()

	return plugin.CallResponse{
		Return: resource.NewPropertyMapFromMap(map[string]interface{}{
			"result": result,
		}),
	}, nil
}
