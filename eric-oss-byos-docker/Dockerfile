ARG BASE_OS_VERSION=5.13.0-14
ARG BASE_OS_URL=armdocker.rnd.ericsson.se/proj-ldc/common_base_os_release/sles
ARG MODULE_NAME

FROM ${BASE_OS_URL}:${BASE_OS_VERSION}

ARG BASE_OS_VERSION
ARG BASE_OS_REPO=arm.sero.gic.ericsson.se/artifactory/proj-ldc-repo-rpm-local/common_base_os/sles

RUN zypper addrepo --gpgcheck-strict -f https://${BASE_OS_REPO}/${BASE_OS_VERSION} baseos \
        && zypper --gpg-auto-import-keys refresh \
        && zypper install -l -y curl \
        && zypper install -l -y wget \
        && zypper clean --all

RUN curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 > get_helm.sh \
        && chmod +x get_helm.sh \
        && ./get_helm.sh
RUN curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
RUN chmod +x ./kubectl
RUN mv ./kubectl /usr/local/bin/kubectl

RUN mkdir /.kube

#COPY kubeconfig /root/.kube/config

#ENV KUBECONFIG=/root/.kube/config
COPY /postgres /postgres
COPY docker_helm.sh /
COPY eric-oss-byos-docker /

LABEL com.ericsson.product-number="CXP 904 3314"
RUN chmod 7777 /postgres
RUN chmod +x docker_helm.sh
#RUN chmod +rwx /root
RUN chmod 7777 /.kube

ARG USER_ID=117518
ARG USER_NAME="eric-oss-restsim-release"

RUN echo "$USER_ID:x:$USER_ID:0:An Identity for $USER_NAME:/nonexistent:/bin/false" >>/etc/passwd
RUN echo "$USER_ID:!::0:::::" >>/etc/shadow


USER $USER_ID

CMD ["/eric-oss-byos-docker"]
