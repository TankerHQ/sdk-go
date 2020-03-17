package core

/*
#cgo CFLAGS: -I${SRCDIR}/ctanker/darwin-amd64/include
#cgo LDFLAGS: -L${SRCDIR}/ctanker/darwin-amd64/lib -lc++ -lc++abi -ltanker_admin-c -lctanker -ltanker_async -ltankerfunctionalhelpers -ltankeradmin -ltankertesthelpers -ltankercore -ltankerstreams -ltankernetwork -ltankertrustchain -ltankeridentity -ltankercrypto -ltankerserialization -ltankererrors -ltankerlog -ltankerformat -ltankercacerts -lsioclient -lsqlpp11-connector-sqlite3 -lfmt -lsodium -ltconcurrent -lboost_program_options -lboost_random -lboost_context -lboost_thread -lboost_chrono -lboost_date_time -lboost_atomic -lboost_filesystem -lboost_system -ldl -lsqlcipher -ltls -lssl -lcrypto
*/
import "C"