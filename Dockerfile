FROM scratch

ADD exec-linux /

ENTRYPOINT ["/exec-linux"]
