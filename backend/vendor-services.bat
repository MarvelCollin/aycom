@echo off
echo Running go mod vendor for all services...

cd %~dp0
echo Vendoring main backend...
go mod vendor

echo Vendoring api-gateway...
cd api-gateway
go mod vendor
cd ..

echo Vendoring event-bus...
cd event-bus
go mod vendor
cd ..

echo Vendoring user service...
cd services\user
go mod vendor
cd ..\..

echo Vendoring thread service...
cd services\thread
go mod vendor
cd ..\..

echo Vendoring community service...
cd services\community
go mod vendor
cd ..\..

echo All services vendored successfully!
