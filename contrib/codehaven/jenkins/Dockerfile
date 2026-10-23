FROM jenkins/jenkins:2.528.1-lts-jdk21

# the init script below configures security, so the interactive setup wizard is not needed
ENV JAVA_OPTS="-Djenkins.install.runSetupWizard=false"

RUN jenkins-plugin-cli --plugins workflow-aggregator git ssh-agent ssh-credentials credentials

COPY init.groovy.d/ /usr/share/jenkins/ref/init.groovy.d/
