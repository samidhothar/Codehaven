// Runs on every Jenkins start: admin account, no anonymous access, public URL and the deployment SSH key
import com.cloudbees.jenkins.plugins.sshcredentials.impl.BasicSSHUserPrivateKey
import com.cloudbees.plugins.credentials.CredentialsScope
import com.cloudbees.plugins.credentials.SystemCredentialsProvider
import com.cloudbees.plugins.credentials.domains.Domain
import hudson.model.User
import hudson.security.FullControlOnceLoggedInAuthorizationStrategy
import hudson.security.HudsonPrivateSecurityRealm
import jenkins.model.Jenkins
import jenkins.model.JenkinsLocationConfiguration

def env = System.getenv()
def jenkins = Jenkins.get()

def adminUser = env['JENKINS_ADMIN_USER'] ?: 'admin'
def adminPassword = env['JENKINS_ADMIN_PASSWORD']
if (!adminPassword) {
  throw new IllegalStateException('JENKINS_ADMIN_PASSWORD must be set')
}
def realm = new HudsonPrivateSecurityRealm(false)
jenkins.setSecurityRealm(realm)
def admin = User.getById(adminUser, false)
if (admin == null) {
  realm.createAccount(adminUser, adminPassword)
} else {
  admin.addProperty(HudsonPrivateSecurityRealm.Details.fromPlainPassword(adminPassword))
}

def authorization = new FullControlOnceLoggedInAuthorizationStrategy()
authorization.setAllowAnonymousRead(false)
jenkins.setAuthorizationStrategy(authorization)

if (env['JENKINS_PUBLIC_URL']) {
  JenkinsLocationConfiguration.get().setUrl(env['JENKINS_PUBLIC_URL'])
}

def keyFile = new File(env['DEPLOY_SSH_KEY_FILE'] ?: '/run/codehaven-keys/id_ed25519')
if (keyFile.exists()) {
  def credentialsID = env['DEPLOY_SSH_CREDENTIALS_ID'] ?: 'codehaven-vm-ssh'
  def key = new BasicSSHUserPrivateKey(CredentialsScope.GLOBAL, credentialsID, 'deploy',
    new BasicSSHUserPrivateKey.DirectEntryPrivateKeySource(keyFile.text), '', 'CodeHaven deployment key')
  def store = SystemCredentialsProvider.getInstance().getStore()
  def existing = store.getCredentials(Domain.global()).find { it.id == credentialsID }
  if (existing) {
    store.updateCredentials(Domain.global(), existing, key)
  } else {
    store.addCredentials(Domain.global(), key)
  }
}

jenkins.setNumExecutors(2)
jenkins.save()
