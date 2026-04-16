# Load the restart_process extension
load('ext://restart_process', 'docker_build_with_restart')

### K8s Config ###
k8s_yaml('./infra/development/k8s/app-config.yaml')
k8s_yaml('./infra/development/k8s/secrets.yaml')
### End of K8s Config ###

### RabbitMQ ###
k8s_yaml('./infra/development/k8s/rabbitmq-deployment.yaml')
k8s_resource('rabbitmq', port_forwards=['5672', '15672'], labels='tooling')
### End RabbitMQ ###

### Redis ###
k8s_yaml('./infra/development/k8s/redis-deployment.yaml')
k8s_resource('redis', port_forwards=['6379'], labels='tooling')
### End Redis ###

### PostgreSQL ###
k8s_yaml('./infra/development/k8s/postgres-deployment.yaml')
k8s_resource('postgres', port_forwards=['5432:5432'], labels='tooling')
### End PostgreSQL ###

### Jaeger ###
k8s_yaml('./infra/development/k8s/jaeger-deployment.yaml')
k8s_resource('jaeger', port_forwards=['16686:16686', '14268:14268'], labels='tooling')
### End Jaeger ###

### Prometheus ###
k8s_yaml('./infra/development/k8s/prometheus-deployment.yaml')
k8s_resource('prometheus', port_forwards=['9091:9090'], labels='tooling')
### End Prometheus ###

### Grafana ###
k8s_yaml('./infra/development/k8s/grafana-deployment.yaml')
k8s_resource('grafana', port_forwards=['3000:3000'], labels='tooling')
### End Grafana ###

### API Service ###
api_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-service ./services/api-service/cmd/main.go'
if os.name == 'nt':
  api_compile_cmd = '.\\infra\\development\\docker\\api-service-build.bat'

local_resource(
  'api-service-compile',
  api_compile_cmd,
  deps=['./services/api-service', './shared'], labels="compiles")

docker_build_with_restart(
  'live-interact/api-service',
  '.',
  entrypoint=['/app/build/api-service'],
  dockerfile='./infra/development/docker/api-service.Dockerfile',
  only=[
    './build/api-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/api-service-deployment.yaml')
k8s_resource('api-service', port_forwards=['8080:8080', '9100:9100'],
             resource_deps=['api-service-compile','rabbitmq', 'redis', 'postgres'], labels="services")
### End of API Service ###

### User Service ###
user_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/user-service ./services/user-service/cmd/main.go'
if os.name == 'nt':
  user_compile_cmd = '.\\infra\\development\\docker\\user-service-build.bat'

local_resource(
  'user-service-compile',
  user_compile_cmd,
  deps=['./services/user-service', './shared'], labels="compiles")

docker_build_with_restart(
  'live-interact/user-service',
  '.',
  entrypoint=['/app/build/user-service'],
  dockerfile='./infra/development/docker/user-service.Dockerfile',
  only=[
    './build/user-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/user-service-deployment.yaml')
k8s_resource('user-service', resource_deps=['user-service-compile','rabbitmq', 'redis', 'postgres'], labels="services")
### End of User Service ###

### Room Service ###
room_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/room-service ./services/room-service/cmd/main.go'
if os.name == 'nt':
  room_compile_cmd = '.\\infra\\development\\docker\\room-service-build.bat'

local_resource(
  'room-service-compile',
  room_compile_cmd,
  deps=['./services/room-service', './shared'], labels="compiles")

docker_build_with_restart(
  'live-interact/room-service',
  '.',
  entrypoint=['/app/build/room-service'],
  dockerfile='./infra/development/docker/room-service.Dockerfile',
  only=[
    './build/room-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/room-service-deployment.yaml')
k8s_resource('room-service', resource_deps=['room-service-compile','rabbitmq', 'redis', 'postgres'], labels="services")
### End of Room Service ###

### Gift Service ###
gift_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/gift-service ./services/gift-service/cmd/main.go'
if os.name == 'nt':
  gift_compile_cmd = '.\\infra\\development\\docker\\gift-service-build.bat'

local_resource(
  'gift-service-compile',
  gift_compile_cmd,
  deps=['./services/gift-service', './shared'], labels="compiles")

docker_build_with_restart(
  'live-interact/gift-service',
  '.',
  entrypoint=['/app/build/gift-service'],
  dockerfile='./infra/development/docker/gift-service.Dockerfile',
  only=[
    './build/gift-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/gift-service-deployment.yaml')
k8s_resource('gift-service', resource_deps=['gift-service-compile','rabbitmq', 'redis', 'postgres'], labels="services")
### End of Gift Service ###

### Danmaku Service ###
danmaku_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/danmaku-service ./services/danmaku-service/cmd/main.go'
if os.name == 'nt':
  danmaku_compile_cmd = '.\\infra\\development\\docker\\danmaku-service-build.bat'

local_resource(
  'danmaku-service-compile',
  danmaku_compile_cmd,
  deps=['./services/danmaku-service', './shared'], labels="compiles")

docker_build_with_restart(
  'live-interact/danmaku-service',
  '.',
  entrypoint=['/app/build/danmaku-service'],
  dockerfile='./infra/development/docker/danmaku-service.Dockerfile',
  only=[
    './build/danmaku-service',
    './shared',
  ],
  live_update=[
    sync('./build', '/app/build'),
    sync('./shared', '/app/shared'),
  ],
)

k8s_yaml('./infra/development/k8s/danmaku-service-deployment.yaml')
k8s_resource('danmaku-service', port_forwards=['9093:9093'],
             resource_deps=['danmaku-service-compile','rabbitmq', 'redis', 'postgres'], labels="services")
### End of Danmaku Service ###
