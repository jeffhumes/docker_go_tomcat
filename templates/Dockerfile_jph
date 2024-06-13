#tomcat image provided from docker.io
FROM gitlab/gitlab-ce

LABEL maintainer="jhumes@ti.com"

ARG UID
ARG GID
ARG USR
ARG IMG_USR_HOME
ARG IMG_USR_SHELL
ARG INSTALL_TYPE

RUN if [[ "${USR}" != "root"  ]]; then groupadd -g "${GID}" "${USR}"; else noop=1; fi
RUN if [[ "${USR}" != "root"  ]]; then useradd -u "${UID}" -g "${GID}" -d ${IMG_USR_HOME} -s ${IMG_USR_SHELL} ${USR}; else noop=1; fi
#COPY if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then  dist/manager-context.xml /usr/local/tomcat/webapps/manager/META-INF/context.xml; else noop=1; fi
#COPY if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then dist/manager-web.xml /usr/local/tomcat/webapps/manager/WEB-INF/web.xml; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then rm -fr /usr/local/tomcat/webapps/examples; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then rm -fr /usr/local/tomcat/webapps/docs; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then mkdir /usr/local/tomcat/properties; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then chown -R ${USR}:${USR} /usr/local/tomcat; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "tomcat"  ]]; then chown -R ${USR}:${USR} /opt/java else; noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "gitlab"  ]]; then chown -R ${USR}:${USR} /var/opt/gitlab; else noop=1; fi
RUN if [[ "${INSTALL_TYPE}" = "gitlab"  ]]; then chown -R ${USR}:${USR} /etc/gitlab; else noop=1; fi
RUN apt-get update 
RUN apt-get -y upgrade 
RUN apt-get -y install sudo
RUN apt-get -y install vim
RUN apt-get -y install git
RUN sed -i 's/%sudo.*$/%sudo  ALL=(ALL) NOPASSWD: ALL/g' /etc/sudoers
RUN usermod -aG sudo ${USR}

#RUN mkdir /usr/local/tomcat/tempcerts
#RUN openssl s_client -connect ubid-prod.itg.ti.com:636 </dev/null | sed -ne '/-BEGIN CERTIFICATE-/,/-END CERTIFICATE-/p' > /usr/local/tomcat/tempcerts/ubid-prod.itg.ti.com.crt; ls -lart /usr/local/tomcat/tempcerts
#RUN cd $JAVA_HOME/conf/security \
#    && keytool -cacerts -storepass changeit -noprompt -trustcacerts -importcert -alias ldapcert -file /usr/local/tomcat/tempcerts/ubid-prod.itg.ti.com.crt

USER ${USR}
