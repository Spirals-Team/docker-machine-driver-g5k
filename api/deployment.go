package api

import (
    "fmt"
    "encoding/json"
    "strings"
    "time"
)

type Deployment struct {
    Nodes   []string    `json:"nodes"`
    Site    string      `json:"site_uid"`
    Status  string      `json:"status"`
    Uid     string      `json:"uid"`
    Links   []Link      `json:"links"`
}

func (a *Api) DeployEnvironment(job *Job) (Job, error) {
    var err error
    urlDeploy := fmt.Sprintf("%s/sites/%s/deployments", G5kApiFrontend, a.Site)

    // Wait for the nodes to be running
    if !a.waitJobIsReady(job) {
        return *job, fmt.Errorf("[G5K_api] Job launching failed");
    }

    // Format arguments
    nodesStrs := make([]string, 0)
    for _, nodes := range job.Nodes {
        nodesStrs = append(nodesStrs, `"` + nodes + `"`)
    }
    nodesJson := strings.Join(nodesStrs, ",")
    fmt.Println(job)

    // Deploying
    deploymentArguments := fmt.Sprintf(`{"nodes": [%s], "environment": %q, "key": "http://public.%s.grid5000.fr/~%s/authorized_keys"}`, nodesJson, defaultImage, a.Site, a.Username)
    var resp []byte
    var deployment Deployment
    fmt.Println(urlDeploy)
    fmt.Println(deploymentArguments)
    resp, err = a.post(urlDeploy, deploymentArguments)
    if err != nil {
        return Job{}, err
    }
    err = json.Unmarshal(resp, &deployment)

    // Waiting the deployment finishes
    fmt.Println(urlDeploy + "/" + deployment.Uid)
    for deployment.Status == "waiting" || deployment.Status == "processing" {
        time.Sleep(10*time.Second)
        resp, err = a.get(urlDeploy + "/" + deployment.Uid)
        if err != nil {
            return Job{}, err
        } else if err = json.Unmarshal(resp, &deployment); err != nil {
            return Job{}, err
        }
        fmt.Println(deployment)
    }
    if deployment.Status != "terminated" {
        return Job{}, fmt.Errorf("[G5K_api] Deployment failed: status is %s\n", deployment.Status)
    }
    return *job, nil
}
