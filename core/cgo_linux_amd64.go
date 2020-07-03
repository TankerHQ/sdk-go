package core

/*
#cgo CFLAGS: -I${SRCDIR}/ctanker/linux-amd64/include
#cgo LDFLAGS: -L${SRCDIR}/ctanker/linux-amd64/lib -ltanker_admin-c -lctanker -ltanker_async -ltankerfunctionalhelpers -ltankeradmin -ltankertesthelpers -ltankercore -ltankerstreams -ltankernetwork -ltankertrustchain -ltankeridentity -ltankercrypto -ltankerserialization -ltankererrors -ltankerlog -ltankerformat -ltankercacerts -lfetchpp -lsioclient -lsqlpp11-connector-sqlite3 -lm -lfmt -lsodium -lpthread -ltconcurrent -lboost_container -lboost_contract -lboost_exception -lboost_program_options -lboost_random -lboost_context -lboost_thread -lboost_chrono -lboost_date_time -lboost_atomic -lboost_filesystem -lboost_system -lboost_stacktrace_addr2line -lboost_stacktrace_backtrace -lboost_stacktrace_basic -lboost_stacktrace_noop -lrt -ldl -lpthread -lsqlcipher -lpthread -ldl -ltls -lssl -lcrypto -lpthread -static-libstdc++ -static-libgcc
*/
import "C"
