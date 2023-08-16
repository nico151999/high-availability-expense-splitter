version: v1
managed:
  enabled: true
  go_package_prefix:
    default: {{ (ds "data").goModule }}/{{ (ds "data").relativeGoLibOutDir }}
    except:
      - buf.build/googleapis/googleapis
      - buf.build/envoyproxy/protoc-gen-validate
      - buf.build/grpc-ecosystem/grpc-gateway
      - buf.build/srikrsna/protoc-gen-gotag
plugins:
  - plugin: buf.build/protocolbuffers/go:v1.30.0
    out: {{ (ds "data").relativeGoLibOutDir }}
    opt:
      - paths=source_relative
  - plugin: buf.build/connectrpc/go:v1.11.0
    out: {{ (ds "data").relativeGoLibOutDir }}
    opt:
      - paths=source_relative
  # we need gRPC for the gateway to work
  - plugin: buf.build/grpc/go:v1.3.0
    out: {{ (ds "data").relativeGoLibOutDir }}
    opt:
      - paths=source_relative
      - require_unimplemented_servers=false
  - plugin: buf.build/grpc-ecosystem/gateway:v2.15.2
    out: {{ (ds "data").relativeGoLibOutDir }}
    opt:
      - paths=source_relative
  - plugin: buf.build/bufbuild/validate-go:v0.10.1
    out: {{ (ds "data").relativeGoLibOutDir }}
    opt:
      - paths=source_relative
  - plugin: buf.build/grpc-ecosystem/openapiv2:v2.15.2
    out: gen/doc
    opt:
      - json_names_for_fields=false
      - proto3_optional_nullable=true
      - allow_merge=true
      - allow_delete_body=true
      - omit_enum_default_value=true
      - output_format=yaml
      - disable_default_responses=true
      - use_go_templates=true
  - plugin: buf.build/community/pseudomuto-doc:v1.5.1
    out: gen/doc
    opt:
      - html
      - index.html
  - plugin: buf.build/bufbuild/es:v1.2.0
    out: gen/lib/ts
    opt:
      - target=ts
  - plugin: buf.build/bufbuild/connect-web:v0.8.6
    out: gen/lib/ts
    opt:
      - target=ts
