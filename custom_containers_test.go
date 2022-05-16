package terralocacon

import (
	"context"
	"testing"
)

func TestLocalstackContainer(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create localstack container",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, externalPort, err := NewLocalstackContainer(tt.args.ctx, "eu-west-1", "dynamodb")
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocalstackContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewLocalstackContainer() got = %v", got)
			}
			if externalPort == "" {
				t.Errorf("NewLocalstackContainer() externalPort = %v", externalPort)
			}
		})
	}
}

func TestNewMongoDBContainer(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create mongo container",
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, externalPort, err := NewMongoDBContainer(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMongoContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewMongoContainer() got = %v", got)
			}
			if externalPort == "" {
				t.Errorf("NewMongoContainer() externalPort = %v", externalPort)
			}
		})
	}
}
