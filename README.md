# Example

api.proto
```api.proto
syntax = "proto3";
package auth.v1;
option go_package = "./;authpb";

message LoginRequest {
    string code = 1;
}

message LoginResponse {
    string access_token = 1;
    int32 expires_in = 2;
}

message UserLoginRequest {
    string user_name = 1;
    string password = 2;
}

service AuthService {
    rpc Login (LoginRequest) returns (LoginResponse);
    rpc UserLogin (UserLoginRequest) returns (LoginResponse);
}
```


```
$ proto-to-dart-http -o . -p flutter_comic -pp server/auth/api example/api.proto
```

```
import 'package:dio/dio.dart';
import 'package:flutter_comic/server/auth/api/api.pb.dart';
class AuthServiceClient {
	final Dio _dio;
	AuthServiceClient(this._dio);

	Future<LoginResponse> login(LoginRequest body, Map<String, dynamic> headers) async {
		Response response = await _dio.post(
			"/AuthService/Login",
			options: Options(headers: headers,),
			data: body.toJson());

		return LoginResponse.fromJson(response.data);
	}

	Future<LoginResponse> userLogin(UserLoginRequest body, Map<String, dynamic> headers) async {
		Response response = await _dio.post(
			"/AuthService/UserLogin",
			options: Options(headers: headers,),
			data: body.toJson());

		return LoginResponse.fromJson(response.data);
	}

}
```


