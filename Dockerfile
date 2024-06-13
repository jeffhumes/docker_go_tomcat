ARG INSTALL_FROM
FROM ${INSTALL_FROM}

LABEL maintainer="jhumes@ti.com"

SHELL ["/bin/bash", "-c"]

ARG UID
ARG GID
ARG USR
ARG IMG_USR_HOME
ARG IMG_USR_SHELL
ARG OS_UPDATES
ARG EXTRA_PACKAGES
ARG INSTALL_TYPE
ARG USE_PROXY
ARG PROXY_URL

#RUN if [[ "${OS_UPDATES}" = "true" ]] || [[ "${EXTRA_PACKAGES}" = "true" ]]; then \
        #if [[ "${USE_PROXY}" = "true"]]; then \
                #apt-get update --option Acquire::HTTP::Proxy=${PROXY_URL} \
        #else \
                #apt-get update \
        #fi \
#fi

#-------------------------------
# OS UPDATES?
#-------------------------------
RUN if [[ "${OS_UPDATES}" = "true" ]] || [[ "${EXTRA_PACKAGES}" = "true" ]] && [[ "${USE_PROXY}" = "true" ]]; then apt-get update --option Acquire::HTTP::Proxy=${PROXY_URL}; fi
RUN if [[ "${OS_UPDATES}" = "true" ]] || [[ "${EXTRA_PACKAGES}" = "true" ]] && [[ "${USE_PROXY}" = "false" ]]; then apt-get update; fi

RUN if [[ "${OS_UPDATES}" = "true" ]] && [[ "${USE_PROXY}" = "true" ]]; then apt-get -y upgrade --option Acquire::HTTP::Proxy=${PROXY_URL}; fi
RUN if [[ "${OS_UPDATES}" = "true" ]] && [[ "${USE_PROXY}" = "false" ]]; then apt-get -y upgrade; fi

RUN if [[ "${EXTRA_PACKAGES}" = "true" ]] && [[ "${USE_PROXY}" = "true" ]]; then apt-get -y install sudo vim git curl net-tools --option Acquire::HTTP::Proxy=${PROXY_URL}; fi
RUN if [[ "${EXTRA_PACKAGES}" = "true" ]] && [[ "${USE_PROXY}" = "false" ]]; then apt-get -y install sudo vim git curl net-tools; fi

RUN if [[ "${EXTRA_PACKAGES}" = "true" ]]; then sed -i 's/%sudo.*$/%sudo  ALL=(ALL) NOPASSWD: ALL/g' /etc/sudoers; fi

RUN if [[ "${USR}" != "root"  ]]; then groupadd -g "${GID}" "${USR}"; else noop=1; fi
RUN if [[ "${USR}" != "root"  ]]; then useradd -u "${UID}" -g "${GID}" -m -d ${IMG_USR_HOME} -s ${IMG_USR_SHELL} ${USR}; else noop=1; fi
RUN if [[ "${USR}" != "root" ]]; then usermod -aG sudo ${USR}; fi

COPY dist/manager-context.xml /tmp/context.xml
COPY dist/manager-web.xml /tmp/web.xml

RUN if [[ "${INSTALL_TYPE}" = "tomcat" ]]; then \
	#cp -av /usr/local/tomcat/webapps.dist/manager /usr/local/tomcat/webapps/ && \
	#cp -av /tmp/context.xml /usr/local/tomcat/webapps/manager/META-INF/context.xml && \
	#cp -av /tmp/web.xml /usr/local/tomcat/webapps/manager/WEB-INF/web.xml && \
        #rm -fr /usr/local/tomcat/webapps/examples && \
        #rm -fr /usr/local/tomcat/webapps/docs && \
        mkdir /usr/local/tomcat/properties && \
        chown -R ${USR}:${USR} /usr/local/tomcat && \
        chown -R ${USR}:${USR} /opt/java; fi

ARG COMPANY

# the following are TI specific, not needed outside TI
RUN if [[ "${INSTALL_TYPE}" = "tomcat" ]] && [[ "${COMPANY}" = "TI" ]]; then \
        mkdir /usr/local/tomcat/tempcerts && \
        openssl s_client -connect ubid-prod.itg.ti.com:636 </dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > /usr/local/tomcat/tempcerts/ubid-prod.itg.ti.com.crt && \
        cd $JAVA_HOME/conf/security && \
        keytool -cacerts -storepass changeit -noprompt -trustcacerts -importcert -alias ldapcert -file /usr/local/tomcat/tempcerts/ubid-prod.itg.ti.com.crt; fi

USER ${USR}
