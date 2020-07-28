package core

/*
#cgo CFLAGS: -I${SRCDIR}/ctanker/windows-386/include
#cgo LDFLAGS: -L${SRCDIR}/ctanker/windows-386/lib -ltanker_admin-c -lctanker -ltanker_async -ltankerfunctionalhelpers -ltankeradmin -ltankertesthelpers -ltankercore -ltankerstreams -ltankernetwork -ltankertrustchain -ltankeridentity -ltankercrypto -ltankerserialization -ltankererrors -ltankerlog -ltankerformat -ltankercacerts -lcrypt32 -lfetchpp -lsioclient -lcrypt32 -lsqlpp11-connector-sqlite3 -lfmt -lsodium -ltconcurrent -lboost_container -lboost_contract -lboost_exception -lboost_program_options -lboost_random -lboost_context -lboost_thread -lboost_chrono -lboost_date_time -lboost_atomic -lboost_filesystem -lboost_system -lboost_stacktrace_noop -lboost_stacktrace_windbg -lws2_32 -lskyr-url -lsqlcipher -ltls-20 -lssl-48 -lcrypto-46 -lws2_32 -static-libstdc++ -static-libgcc
*/
import "C"
