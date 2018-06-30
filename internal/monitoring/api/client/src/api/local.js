const fetchImages = () => new Promise((resolve) => {
  setTimeout(() => {
    resolve(
      [{"id":"34056f1481eef17bdae484fcd900379d8a17528293dd0d78e5474ba191a630fd","tag":"docker.io/anchore/anchore-engine:latest","owner":["virity"],"cve_count":4},{"id":"969f714dc41e6cdff0bb22f3a09eaef72b1717139cdb9b9b42434eeba65dbb1a","tag":"hiroakis/docker-sensu-server","owner":["virity"],"cve_count":3},{"id":"65bf726222e13b0ceff0bb20bb6f0e99cbf403a7a1f611fdd2aadd0c8919bbcf","tag":"postgres:10","owner":["virity"],"cve_count":19},{"id":"61ad638751093d94c7878b17eee862348aa9fc5b705419b805f506d51b9882e7","tag":"quay.io/coreos/etcd:v3.3","owner":["virity"],"cve_count":4},{"id":"868c71bbeac05b0265afc25a1fef0b531b0829bb96d5d6e6faee55bef751dfd9","tag":"kaitsh/sensu-client-testing","owner":["virity"],"cve_count":22}]
  )}, 2000);
});

const fetchImageDetails = (id) => {
  if (id == "61ad638751093d94c7878b17eee862348aa9fc5b705419b805f506d51b9882e7") {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(
          {"id":"61ad638751093d94c7878b17eee862348aa9fc5b705419b805f506d51b9882e7","tag":"quay.io/coreos/etcd:v3.3","in_containers":[{"Name":"/virity-example_etcd_1","ID":"18502a9abeb87709bc60105a125d12abe22bcbfe6f784568928cf907f8ff84c1","Hostname":"demo-host","Image":"quay.io/coreos/etcd:v3.3","ImageID":"61ad638751093d94c7878b17eee862348aa9fc5b705419b805f506d51b9882e7","OwnerID":"virity","Timestamp":"2018-06-29T16:45:31Z"}],"vulnerability_cve":[{"Fix":"2.6.5-r0","Package":"libressl2.6-libcrypto-2.6.3-r0","Severity":0,"URL":"http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-0732","Vuln":"CVE-2018-0732","Description":""},{"Fix":"2.6.5-r0","Package":"libressl2.6-libssl-2.6.3-r0","Severity":0,"URL":"http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-0495","Vuln":"CVE-2018-0495","Description":""},{"Fix":"2.6.5-r0","Package":"libressl2.6-libssl-2.6.3-r0","Severity":0,"URL":"http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-0732","Vuln":"CVE-2018-0732","Description":""},{"Fix":"2.6.5-r0","Package":"libressl2.6-libcrypto-2.6.3-r0","Severity":0,"URL":"http://cve.mitre.org/cgi-bin/cvename.cgi?name=CVE-2018-0495","Vuln":"CVE-2018-0495","Description":""}],"owner":["virity"]}
      )}, 2000)
    })
  } else if (id == "969f714dc41e6cdff0bb22f3a09eaef72b1717139cdb9b9b42434eeba65dbb1a") {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(
          {"id":"969f714dc41e6cdff0bb22f3a09eaef72b1717139cdb9b9b42434eeba65dbb1a","tag":"hiroakis/docker-sensu-server","in_containers":[{"Name":"/virity-example_sensu-server_1","ID":"c4d779d4555a99cbbffec7c25160ba92e6f0fd1b44be1b95d46917983188cd75","Hostname":"demo-host","Image":"hiroakis/docker-sensu-server","ImageID":"969f714dc41e6cdff0bb22f3a09eaef72b1717139cdb9b9b42434eeba65dbb1a","OwnerID":"virity","Timestamp":"2018-06-29T16:45:31Z"}],"vulnerability_cve":[{"Fix":"0:2.12-1.212.el6","Package":"glibc-2.12-1.149.el6_6.9","Severity":0,"URL":"https://access.redhat.com/errata/RHSA-2018:1879","Vuln":"RHSA-2018:1879","Description":""},{"Fix":"0:2.12-1.212.el6","Package":"glibc-common-2.12-1.149.el6_6.9","Severity":0,"URL":"https://access.redhat.com/errata/RHSA-2018:1879","Vuln":"RHSA-2018:1879","Description":""},{"Fix":"0:3.2.8-45.el6_9.3","Package":"procps-3.2.8-30.el6","Severity":0,"URL":"https://access.redhat.com/errata/RHSA-2018:1777","Vuln":"RHSA-2018:1777","Description":""}],"owner":["virity"]}
      )}, 2000)
    })
  }
}


export default {
  fetchImages,
  fetchImageDetails
}