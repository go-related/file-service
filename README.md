# File Service
File service is an application that will load a big json file and <br/> 
will send it to a grpc server to save to the database.
This will be a hexadecimal architecture implementation.



## Notes
- To build the proto - we will put the 
  ```shell 
  protoc --go_out=./proto --go-grpc_out=./proto proto/file.proto
    ```

