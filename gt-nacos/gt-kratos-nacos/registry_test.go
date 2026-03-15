package gtknacos

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	gtnacos "github.com/DontBeProud/go-treasure/gt-nacos"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
	"github.com/go-kratos/kratos/v2/registry"
)

var (
	testIp = "your_ip"
	nc     = &gtconfpb.NacosConfig{
		Endpoint:  []string{testIp + ":8848"},
		Namespace: "your_namespace",
		UserName:  "your_username",
		Password:  "your_password",
		AccessKey: "",
		SecretKey: "",
		LogDir:    nil,
		CacheDir:  nil,
		Timeout:   nil,
	}
	nrc = &gtconfpb.NacosRegistryConfig{
		Group:   "",
		Prefix:  "",
		Weight:  0,
		Cluster: "",
		Kind:    "",
	}
)

// skipTest 跳过测试，判断的依据是环境变量是否有配置 RUN_TEST=true
func skipTest(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_TEST") != "true" {
		t.Skip("skipping test; set RUN_TEST=true to run")
	}
}

func TestRegistry_Register(t *testing.T) {
	skipTest(t)
	c, _, _ := gtnacos.NewClient(nc)
	rc, _, _ := c.NewRegistryCenterClient()
	regClient := rc.GetClient()
	kratosRegClient := NewKratosRegistryClient(rc, nrc)
	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test1",
		Version:   "v1.0.0",
		Endpoints: []string{"http://" + testIp + ":8080?isSecure=false"},
	}
	testServerWithMetadata := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test1",
		Version:   "v1.0.0",
		Endpoints: []string{"http://" + testIp + ":8080?isSecure=false"},
		Metadata:  map[string]string{"idc": "shanghai-xs"},
	}
	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx     context.Context
		service *registry.ServiceInstance
	}
	var err error
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		deferFunc func(t *testing.T)
	}{
		{
			name: "normal",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
			deferFunc: func(t *testing.T) {
				err = kratosRegClient.Deregister(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "withMetadata",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx:     context.Background(),
				service: testServerWithMetadata,
			},
			wantErr: false,
			deferFunc: func(t *testing.T) {
				err = kratosRegClient.Deregister(context.Background(), testServerWithMetadata)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "error",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "",
					Version:   "v1.0.0",
					Endpoints: []string{"http://" + testIp + ":8080?isSecure=false"},
				},
			},
			wantErr: true,
		},
		{
			name: "urlError",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "test",
					Version:   "v1.0.0",
					Endpoints: []string{testIp + ":8080"},
				},
			},
			wantErr: true,
		},
		{
			name: "portError",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx: context.Background(),
				service: &registry.ServiceInstance{
					ID:        "1",
					Name:      "test",
					Version:   "v1.0.0",
					Endpoints: []string{"http://" + testIp + "888"},
				},
			},
			wantErr: true,
		},
		{
			name: "withCluster",
			fields: fields{
				registry: New(regClient, WithCluster("test")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withGroup",
			fields: fields{
				registry: New(regClient, WithGroup("TEST_GROUP")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withWeight",
			fields: fields{
				registry: New(regClient, WithWeight(200)),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
		{
			name: "withPrefix",
			fields: fields{
				registry: New(regClient, WithPrefix("test")),
			},
			args: args{
				ctx:     context.Background(),
				service: testServer,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.registry
			if err := r.Register(tt.args.ctx, tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("Register error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRegistry_GetService(t *testing.T) {
	skipTest(t)
	c, _, _ := gtnacos.NewClient(nc)
	rc, _, _ := c.NewRegistryCenterClient()
	kratosRegClient := NewKratosRegistryClient(rc, nrc)
	var err error
	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test3",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://" + testIp + ":8080?isSecure=false"},
	}

	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx         context.Context
		serviceName string
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		want      []*registry.ServiceInstance
		wantErr   bool
		preFunc   func(t *testing.T)
		deferFunc func(t *testing.T)
	}{
		{
			name: "normal",
			preFunc: func(t *testing.T) {
				err = kratosRegClient.Register(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
				time.Sleep(time.Second * 3)
			},
			deferFunc: func(t *testing.T) {
				err = kratosRegClient.Deregister(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
			fields: fields{
				registry: kratosRegClient,
			},
			args: args{
				ctx:         context.Background(),
				serviceName: testServer.Name + "." + "grpc",
			},
			want: []*registry.ServiceInstance{{
				ID:        testIp + "#8080#DEFAULT#DEFAULT_GROUP@@test3.grpc",
				Name:      "DEFAULT_GROUP@@test3.grpc",
				Version:   "v1.0.0",
				Metadata:  map[string]string{"version": "v1.0.0", "kind": "grpc", "weight": "100"},
				Endpoints: []string{"grpc://" + testIp + ":8080"},
			}},
			wantErr: false,
		},
		{
			name: "errorNotExist",
			fields: fields{
				registry: kratosRegClient,
			},
			args: args{
				ctx:         context.Background(),
				serviceName: "notExist",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preFunc != nil {
				tt.preFunc(t)
			}
			if tt.deferFunc != nil {
				defer tt.deferFunc(t)
			}
			r := tt.fields.registry
			got, err := r.GetService(tt.args.ctx, tt.args.serviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetService error = %v, wantErr %v", err, tt.wantErr)
				t.Errorf("GetService got = %v", got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetService got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistry_Watch(t *testing.T) {
	skipTest(t)
	c, _, _ := gtnacos.NewClient(nc)
	rc, _, _ := c.NewRegistryCenterClient()
	regClient := rc.GetClient()
	kratosRegClient := NewKratosRegistryClient(rc, nrc)
	testServer := &registry.ServiceInstance{
		ID:        "1",
		Name:      "test4",
		Version:   "v1.0.0",
		Endpoints: []string{"grpc://" + testIp + ":8080?isSecure=false"},
	}

	cancelCtx, cancel := context.WithCancel(context.Background())
	type fields struct {
		registry *Registry
	}
	type args struct {
		ctx         context.Context
		serviceName string
	}
	var err error
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		want        []*registry.ServiceInstance
		processFunc func(t *testing.T)
	}{
		{
			name: "normal",
			fields: fields{
				registry: New(regClient),
			},
			args: args{
				ctx:         context.Background(),
				serviceName: testServer.Name + "." + "grpc",
			},
			wantErr: false,
			want: []*registry.ServiceInstance{{
				ID:        testIp + "#8080#DEFAULT#DEFAULT_GROUP@@test4.grpc",
				Name:      "DEFAULT_GROUP@@test4.grpc",
				Version:   "v1.0.0",
				Metadata:  map[string]string{"version": "v1.0.0", "kind": "grpc"},
				Endpoints: []string{"grpc://" + testIp + ":8080"},
			}},
			processFunc: func(t *testing.T) {
				err = kratosRegClient.Register(context.Background(), testServer)
				if err != nil {
					t.Error(err)
				}
			},
		},
		{
			name: "ctxCancel",
			fields: fields{
				registry: kratosRegClient,
			},
			args: args{
				ctx:         cancelCtx,
				serviceName: testServer.Name,
			},
			wantErr: true,
			want:    nil,
			processFunc: func(*testing.T) {
				cancel()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.fields.registry
			watch, err := r.Watch(tt.args.ctx, tt.args.serviceName)
			if err != nil {
				t.Error(err)
				return
			}
			defer func() {
				err = watch.Stop()
				if err != nil {
					t.Error(err)
				}
			}()
			_, err = watch.Next()
			if err != nil {
				t.Error(err)
				return
			}

			if tt.processFunc != nil {
				tt.processFunc(t)
			}

			want, err := watch.Next()
			if (err != nil) != tt.wantErr {
				t.Errorf("Watch error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(want, tt.want) {
				t.Errorf("Watch watcher = %v, want %v", watch, tt.want)
			}
		})
	}
}
