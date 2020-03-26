package core

/*
#cgo CFLAGS: -I${SRCDIR}/ctanker/linux-amd64/include
#cgo LDFLAGS: -L${SRCDIR}/ctanker/linux-amd64/lib -ltanker_admin-c -lctanker -ltanker_async -ltankerfunctionalhelpers -ltankeradmin -ltankertesthelpers -ltankercore -ltankerstreams -ltankernetwork -ltankertrustchain -ltankeridentity -ltankercrypto -ltankerserialization -ltankererrors -ltankerlog -ltankerformat -ltankercacerts -lsioclient -lsqlpp11-connector-sqlite3 -lfmt -lsodium -ltconcurrent -lboost_program_options -lboost_random -lboost_context -lboost_thread -lboost_chrono -lboost_date_time -lboost_atomic -lboost_filesystem -lboost_system -lpthread -ldl -lsqlcipher -lpthread -ldl -ltls -lssl -lcrypto -lpthread -static-libstdc++ -static-libgcc
*/
import "C"
