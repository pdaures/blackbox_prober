FROM scratch
MAINTAINER Patrick Daures <patrick.daures@gmail.com>
ADD blackbox_prober_linux_amd64 /blackbox_prober_linux_amd64
EXPOSE 9115
CMD ["/blackbox_prober_linux_amd64"]


