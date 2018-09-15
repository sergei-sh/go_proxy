Simple Golang caching proxy server

Features:
1. Works as explicit HTTP/HTTPS proxy
2. HTTP responses are passed from origin. If a picture requested it is saved
to a database (PostgreSQL). If already saved picture is requested it is sent from cache.
3. DB writes are pipelined and performed in their separate routine (no concurrent writes)
4. HTTPS connections are being tunnelled
5. Upon pressing Ctrl-C requestes are waited to be completed and DB jobs are waited to be done,
then the server is closed.

Installation/Prerequisites:
1.go_proxy source
2.lib/pq PostgreSQL DB Go driver 
3.PostgreSQL DB server running
4.Some browser and internet connection

Usage (suppose using Chromium browser):
1. Setup desired proxy port and DB addres/credentials in the head of go_proxy.go file 
2. go install
3. ./go_proxy 
4. chromium-browser --proxy-server=http://localhost:8066
5. Clear cached pictures: Menu -> Settings -> Search: type in "Clear" -> Clear browsing data -> Check Cached images and files -> Clear browsing data
6. Use baidu com to test HTTP with pictures:
    http://www.baidu.com/s?ie=utf-8&f=8&rsv_bp=1&rsv_idx=1&tn=baidu&wd=pictures&oq=pictures
7. Use any other website to test HTTPS
