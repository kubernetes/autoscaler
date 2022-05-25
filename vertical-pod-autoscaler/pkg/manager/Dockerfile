FROM alpine:3.13.6
ADD ./bin/vpa-informer /informer
RUN chmod +x ./informer
ENTRYPOINT ["./informer"]
CMD ["--default-update-mode", "Initial"]
