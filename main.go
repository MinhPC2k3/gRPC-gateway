package main

import (
	//"github.com/gin-gonic/gin"
	helloworldpb "api1/gen/go/hello"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// command-line options:
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("server-port", "8091", "gRPC server port")
	grpcGatewayPort    = flag.String("gateway-port", "8090", "Gateway port")
)

func OpenDB() *gorm.DB {
	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/ShangHai"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Account{})
	return db
}

type server struct {
	helloworldpb.UnimplementedGreeterServer
}

func NewServer() *server {
	return &server{}
}
func (s *server) SayHello(ctx context.Context, in *helloworldpb.HelloRequest) (*helloworldpb.HelloReply, error) {
	return &helloworldpb.HelloReply{Message: in.Name}, nil
}
func (s *server) GetData(ctx context.Context, in *helloworldpb.Empty) (*helloworldpb.Accounts, error) {
	db := OpenDB()
	var test Account
	db.Find(&test).Scan(&getData)

	var listAcc []*helloworldpb.Account

	for _, acc := range getData {
		listAcc = append(listAcc, &helloworldpb.Account{
			Name:    acc.Name,
			Gender:  acc.Gender,
			Address: acc.Address,
		})
	}

	return &helloworldpb.Accounts{
		Minh: listAcc,
	}, nil
}
func (s *server) CreateAcc(ctx context.Context, in *helloworldpb.Account) (*helloworldpb.NotifyReply, error) {
	db := OpenDB()
	var newAccount Account
	newAccount = Account{Name: "nguyen van a", Gender: "male", Address: "Hanoi"}
	tm := db.Create(&newAccount)
	if tm.Error != nil {
		return &helloworldpb.NotifyReply{
			Message: fmt.Sprintf("Create Account %s error at %v", in.GetName(), tm.Error),
		}, nil
	}
	return &helloworldpb.NotifyReply{
		Message: fmt.Sprintf("Create Account %s to database success", in.GetName()),
	}, nil
}

type Account struct {
	gorm.Model
	Name    string `json:"name"`
	Gender  string `json:"gender"`
	Address string `json:"address"`
}

var getData = []Account{}

func main() {
	flag.Parse()

	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/ShangHai"
	db, err1 := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err1 != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&Account{})
	err := establishGrpcServer()

	if err != nil {
		log.Fatal(err)
	}

	err = establisGrpcGateway()

	if err != nil {
		log.Fatal(err)
	}

}
func establishGrpcServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", *grpcServerEndpoint))
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	s := grpc.NewServer()
	helloworldpb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	go func() {
		log.Fatalln(s.Serve(lis))
	}()
	return nil
}

func establisGrpcGateway() error {
	conn, err := grpc.DialContext(
		context.Background(),
		fmt.Sprintf("0.0.0.0:%s", *grpcServerEndpoint),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register Greeter
	err = helloworldpb.RegisterGreeterHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", *grpcGatewayPort),
		Handler: gwmux,
	}

	log.Printf("Serving gRPC-Gateway on http://0.0.0.0:%s\n", *grpcGatewayPort)
	log.Fatalln(gwServer.ListenAndServe())

	return nil
}

// func main() {
// 	flag.Parse()

// 	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/ShangHai"
// 	db,err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}
// 	db.AutoMigrate(&Account{})
// 	err := establishGrpcServer()

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = establisGrpcGateway()

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }
// func addAccount(context *gin.Context) {
// 	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
// 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}
// 	db.AutoMigrate(&Account{})
// 	var newAccount Account
// 	if err := context.BindJSON(&newAccount); err != nil {
// 		return
// 	}
// 	db.Select("ID", "Name", "Gender", "Address").Create(&newAccount)
// 	// db.Find(&Accounts)
// 	context.IndentedJSON(http.StatusCreated, newAccount)
// }
// func getAccount(context * gin.Context) {
// 	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/ShangHai"
// 	db,err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		panic("failed to connect database")
// 	}
// 	db.AutoMigrate(&Account{})
// 	var test Account
// 	db.Find(&test).Scan(&getData)
// 	context.IndentedJSON(http.StatusOK, getData)
// }

// func main() {
// 	router := gin.Default()
// 	router.GET("/Accounts", getAccount)
// 	router.POST("/Accounts", addAccount)
// 	router.Run("localhost:8080")

// }

// // var DB *gorm.DB
// // var err error
// // func InitialMigration(){
// // 	dsn := "host=localhost user=gorm password=gorm dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
// // 	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// //   if err != nil {
// //     panic("failed to connect database")
// //   }
// //   db.AutoMigrate(&Account{})
// // }
// package main

// import (
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// 	"fmt"
// )

// func main() {
// 	type Account struct {
// 		gorm.Model
// 		Name    string `json:"name"`
// 		Gender  string `json:"gender"`
// 		Address string `json:"address"`
// 	}
// 	getData:=[]Account{}
// 	dsn := "host=localhost user=gorm password=grom dbname=gorm port=5432 sslmode=disable TimeZone=Asia/Shanghai"
// 	db,err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if(err!= nil){
// 		panic("Failed to connect to database")
// 	}
// 	db.AutoMigrate(&Account{})
// 	//var accounts Account
// 	A:=Account{}
// 	db.Find(&A).Scan(&getData)
// 	for _,x := range getData{
// 		fmt.Println(x);
// 	}
// }
