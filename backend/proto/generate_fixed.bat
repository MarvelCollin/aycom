@echo off
echo Generating protocol buffer files...
cd ..
protoc --go_out=. --go_opt=paths=source_relative ^
       --go-grpc_out=. --go-grpc_opt=paths=source_relative ^
       proto/auth/auth.proto proto/community/community.proto proto/thread/thread.proto proto/user/user.proto
echo Done! 