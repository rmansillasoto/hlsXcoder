@Library('Utils') _

def imageName = 'hlstranscoder'
def dockerFile = 'Dockerfile'

node('docker-builder')
{
    try{
        stage('Checkout SCM')
        {
            checkoutResults = checkout scm
            repoVersion = git.GetRepositoryVersionAndTagIt(checkoutResults)
        }
        if(BRANCH_NAME == "master" || BRANCH_NAME == "develop")
        {
            stage('Build Image')
            {
                dockerFactory.Build(imageName, "--no-cache -f ${env.WORKSPACE}/${dockerFile} ${env.WORKSPACE}")
            }
            stage('Push to registry')
            {
                if(BRANCH_NAME == "master")
                {
                    dockerFactory.PushWithTag(imageName, "${imageName}:${repoVersion}")
                    manager.addShortText("${repoVersion}")
                    dockerFactory.PushWithTag(imageName,"${imageName}:latest")
                }
                else if (BRANCH_NAME == "develop")
                {
                    dockerFactory.PushWithTag(imageName, "${imageName}:dev-${repoVersion}")
                    manager.addShortText("${repoVersion}")
                    dockerFactory.PushWithTag(imageName,"${imageName}:dev-latest")
                }
            }
        }
        else
        {
            log.Info("Branch \'${BRANCH_NAME}\' not eligible for CI/CD purposes")
        }
    }
    catch(all)
    {
        job.ErrorWithEmail("rmansillasoto@gmail.com", "Build failed, Unknown Error")
        
    }
    finally
    {
        try
        {
            dockerFactory.RemoveImage("${imageName}")
        }
        catch(all)
        {
            log.Warning("Omitting docker rmi")
        }

        workspace.Clean()
    }
    
}