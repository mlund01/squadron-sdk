package squadron

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	pb "github.com/mlund01/squadron-sdk/proto"
)

var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "SQUAD_PLUGIN",
	MagicCookieValue: "squadron-tool-plugin-v1",
}

type ToolInfo struct {
	Name         string
	Description  string
	Schema       Schema
	RawSchema    json.RawMessage
	OutputSchema json.RawMessage
}

type ToolProvider interface {
	Configure(settings map[string]string) error
	Call(ctx context.Context, toolName string, payload string) (string, error)
	GetToolInfo(toolName string) (*ToolInfo, error)
	ListTools() ([]*ToolInfo, error)
}

type ToolPluginGRPCPlugin struct {
	plugin.Plugin
	Impl ToolProvider
}

func (p *ToolPluginGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	pb.RegisterToolPluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *ToolPluginGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: pb.NewToolPluginClient(c)}, nil
}

type GRPCClient struct {
	client pb.ToolPluginClient
}

func (c *GRPCClient) Configure(settings map[string]string) error {
	resp, err := c.client.Configure(context.Background(), &pb.ConfigureRequest{
		Settings: settings,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("configure failed: %s", resp.Error)
	}
	return nil
}

func (c *GRPCClient) Call(ctx context.Context, toolName string, payload string) (string, error) {
	resp, err := c.client.Call(ctx, &pb.CallRequest{
		ToolName: toolName,
		Payload:  payload,
	})
	if err != nil {
		return "", err
	}
	return resp.Result, nil
}

func (c *GRPCClient) GetToolInfo(toolName string) (*ToolInfo, error) {
	resp, err := c.client.GetToolInfo(context.Background(), &pb.GetToolInfoRequest{
		ToolName: toolName,
	})
	if err != nil {
		return nil, err
	}
	return protoToToolInfo(resp.Tool)
}

func (c *GRPCClient) ListTools() ([]*ToolInfo, error) {
	resp, err := c.client.ListTools(context.Background(), &pb.ListToolsRequest{})
	if err != nil {
		return nil, err
	}
	tools := make([]*ToolInfo, 0, len(resp.Tools))
	for _, t := range resp.Tools {
		info, err := protoToToolInfo(t)
		if err != nil {
			return nil, err
		}
		tools = append(tools, info)
	}
	return tools, nil
}

func protoToToolInfo(t *pb.ToolInfo) (*ToolInfo, error) {
	info := &ToolInfo{
		Name:        t.Name,
		Description: t.Description,
	}
	if t.SchemaJson != "" {
		raw := json.RawMessage(t.SchemaJson)
		info.RawSchema = raw
		var schema Schema
		_ = json.Unmarshal(raw, &schema)
		info.Schema = schema
	}
	if t.OutputSchemaJson != "" {
		info.OutputSchema = json.RawMessage(t.OutputSchemaJson)
	}
	return info, nil
}

type GRPCServer struct {
	pb.UnimplementedToolPluginServer
	Impl ToolProvider
}

func (s *GRPCServer) Configure(ctx context.Context, req *pb.ConfigureRequest) (*pb.ConfigureResponse, error) {
	err := s.Impl.Configure(req.Settings)
	if err != nil {
		return &pb.ConfigureResponse{Success: false, Error: err.Error()}, nil
	}
	return &pb.ConfigureResponse{Success: true}, nil
}

func (s *GRPCServer) Call(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error) {
	result, err := s.Impl.Call(ctx, req.ToolName, req.Payload)
	if err != nil {
		return nil, err
	}
	return &pb.CallResponse{Result: result}, nil
}

func (s *GRPCServer) GetToolInfo(ctx context.Context, req *pb.GetToolInfoRequest) (*pb.GetToolInfoResponse, error) {
	info, err := s.Impl.GetToolInfo(req.ToolName)
	if err != nil {
		return nil, err
	}
	return &pb.GetToolInfoResponse{
		Tool: toolInfoToProto(info),
	}, nil
}

func (s *GRPCServer) ListTools(ctx context.Context, req *pb.ListToolsRequest) (*pb.ListToolsResponse, error) {
	tools, err := s.Impl.ListTools()
	if err != nil {
		return nil, err
	}
	protoTools := make([]*pb.ToolInfo, 0, len(tools))
	for _, t := range tools {
		protoTools = append(protoTools, toolInfoToProto(t))
	}
	return &pb.ListToolsResponse{Tools: protoTools}, nil
}

func toolInfoToProto(t *ToolInfo) *pb.ToolInfo {
	var schemaJSON []byte
	if len(t.RawSchema) > 0 {
		schemaJSON = t.RawSchema
	} else {
		schemaJSON, _ = json.Marshal(t.Schema)
	}
	return &pb.ToolInfo{
		Name:             t.Name,
		Description:      t.Description,
		SchemaJson:       string(schemaJSON),
		OutputSchemaJson: string(t.OutputSchema),
	}
}

var PluginMap = map[string]plugin.Plugin{
	"tool": &ToolPluginGRPCPlugin{},
}
