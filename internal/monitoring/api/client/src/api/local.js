const fetchImages = () => new Promise((resolve) => {
  setTimeout(() => {
    resolve([
      {"id":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","tag":"debian:latest","in_containers":[{"Name":"/test","ID":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","Hostname":"localhost","Image":"debian:latest","ImageID":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","OwnerID":"","Timestamp":"0001-01-01T00:00:00Z"}],"vulnerability_cve":[{"Fix":"None","Package":"tar-1.29b-1.1","Severity":3,"URL":"https://security-tracker.debian.org/tracker/CVE-2005-2541","Vuln":"CVE-2005-2541","Description":""},{"Fix":"0.0.9+deb9u1","Package":"sensible-utils-0.0.9","Severity":1,"URL":"https://security-tracker.debian.org/tracker/CVE-2017-17512","Vuln":"CVE-2017-17512","Description":""},{"Fix":"None","Package":"util-linux-2.29.2-1","Severity":0,"URL":"https://security-tracker.debian.org/tracker/CVE-2016-2779","Vuln":"CVE-2016-2779","Description":""},{"Fix":"None","Package":"apt-1.4.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""},{"Fix":"YES","Package":"apt-1.9.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""}],"owner":["VIRITY"]},{"id":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","tag":"ubuntu:latest","in_containers":[{"Name":"/test","ID":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","Hostname":"localhost","Image":"debian:latest","ImageID":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","OwnerID":"","Timestamp":"0001-01-01T00:00:00Z"}],"vulnerability_cve":[{"Fix":"None","Package":"tar-1.29b-1.1","Severity":3,"URL":"https://security-tracker.debian.org/tracker/CVE-2005-2541","Vuln":"CVE-2005-2541","Description":""},{"Fix":"0.0.9+deb9u1","Package":"sensible-utils-0.0.9","Severity":1,"URL":"https://security-tracker.debian.org/tracker/CVE-2017-17512","Vuln":"CVE-2017-17512","Description":""},{"Fix":"None","Package":"util-linux-2.29.2-1","Severity":0,"URL":"https://security-tracker.debian.org/tracker/CVE-2016-2779","Vuln":"CVE-2016-2779","Description":""},{"Fix":"None","Package":"apt-1.4.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""},{"Fix":"YES","Package":"apt-1.9.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""}],"owner":["VIRITY"]}
    ]
  )}, 1000);
});

const fetchImageDetails = (id) => {
  if (id == "da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e") {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(
          {"id":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","tag":"debian:latest","in_containers":[{"Name":"/test","ID":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","Hostname":"localhost","Image":"debian:latest","ImageID":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","OwnerID":"","Timestamp":"0001-01-01T00:00:00Z"}],"vulnerability_cve":[{"Fix":"None","Package":"tar-1.29b-1.1","Severity":3,"URL":"https://security-tracker.debian.org/tracker/CVE-2005-2541","Vuln":"CVE-2005-2541","Description":""},{"Fix":"0.0.9+deb9u1","Package":"sensible-utils-0.0.9","Severity":1,"URL":"https://security-tracker.debian.org/tracker/CVE-2017-17512","Vuln":"CVE-2017-17512","Description":""},{"Fix":"None","Package":"util-linux-2.29.2-1","Severity":0,"URL":"https://security-tracker.debian.org/tracker/CVE-2016-2779","Vuln":"CVE-2016-2779","Description":""},{"Fix":"None","Package":"apt-1.4.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""},{"Fix":"YES","Package":"apt-1.9.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""}],"owner":["VIRITY"]}
      )}, 1000)
    })
  } else if (id == "bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800") {
    return new Promise((resolve) => {
      setTimeout(() => {
        resolve(
          {"id":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","tag":"ubuntu:latest","in_containers":[{"Name":"/test","ID":"bf51c9974229f0a3790366464fef13e2cdbf0be5b682874f4e78f1538005a800","Hostname":"localhost","Image":"debian:latest","ImageID":"da653cee0545dfbe3c1864ab3ce782805603356a9cc712acc7b3100d9932fa5e","OwnerID":"","Timestamp":"0001-01-01T00:00:00Z"}],"vulnerability_cve":[{"Fix":"None","Package":"tar-1.29b-1.1","Severity":3,"URL":"https://security-tracker.debian.org/tracker/CVE-2005-2541","Vuln":"CVE-2005-2541","Description":""},{"Fix":"0.0.9+deb9u1","Package":"sensible-utils-0.0.9","Severity":1,"URL":"https://security-tracker.debian.org/tracker/CVE-2017-17512","Vuln":"CVE-2017-17512","Description":""},{"Fix":"None","Package":"util-linux-2.29.2-1","Severity":0,"URL":"https://security-tracker.debian.org/tracker/CVE-2016-2779","Vuln":"CVE-2016-2779","Description":""},{"Fix":"None","Package":"apt-1.4.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""},{"Fix":"YES","Package":"apt-1.9.8","Severity":2,"URL":"https://security-tracker.debian.org/tracker/CVE-2011-3374","Vuln":"CVE-2011-3374","Description":""}],"owner":["VIRITY"]}
      )}, 1000)
    })
  }
}


export default {
  fetchImages,
  fetchImageDetails
}