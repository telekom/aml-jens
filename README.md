<h1 align="center">
    JENS - a tool to simulate L4S marking of a Baseband Unit
</h1>

<p align="center">
    <a href="/../../commits/" title="Last Commit"><img src="https://img.shields.io/github/last-commit/telekom/aml-jens?style=flat"></a>
    <a href="/../../issues" title="Open Issues"><img src="https://img.shields.io/github/issues/telekom/aml-jens?style=flat"></a>
    <a href="./LICENSE" title="License"><img src="https://img.shields.io/badge/License-Apache%202.0-green.svg?style=flat"></a>
</p>

<p align="center">
  <a href="#development">Development</a> •
  <a href="#documentation">Documentation</a> •
  <a href="#support-and-feedback">Support</a> •
  <a href="#how-to-contribute">Contribute</a> •
  <a href="#licensing">Licensing</a>
</p>

This repository contains a go-package for playing & displaying a so called 'data rate pattern' (DRP) 
on a network interface which leverages a l4s capable queue and a custom version of the iproute2 package 
to simulate the marking behavior for one User Equipment (UE) of a baseband unit (BBU).
Also included are three programs, that enable a user to use this functionality through the CLI.

The DRP is defined in a csv file, an example is provided.


Using the `drplay` utility the link-capacity is changed over time in uplink direction in a fixed time period (frequency up to 100 Hz) using the values specified in the DRP.
Measures of the state of the L4S queue are sampled (10ms) and can be persisted (csv or psql). 

`drbenchmarks` enables repetitive calls of drplay. A benchmark is specified as a JSON file.

The command `drshow` visualizes measures or data rate patterns on a terminal ui.

## Support and Feedback

The following channels are available for discussions, feedback, and support requests:

| Type               | Channel                                                                                                                                                                                            |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Issues**         | <a href="/../../issues/new/choose" title="General Discussion"><img src="https://img.shields.io/github/issues/telekom/aml-jens?style=flat-square"></a> </a>                                       |
| **Other Requests** | <a href="mailto:opensource@telekom.de" title="Email Open Source Team"><img src="https://img.shields.io/badge/email-Open%20Source%20Team-green?logo=mail.ru&style=flat-square&logoColor=white"></a> |

## How to Contribute

Contribution and feedback is encouraged and always welcome. For more information about how to contribute, the project structure, as well as additional contribution information, see our [Contribution Guidelines](./CONTRIBUTING.md). 

This project has adopted the [Contributor Covenant](https://www.contributor-covenant.org/) in version 2.1 as our code of conduct. Please see the details in our [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md). All contributors must abide by the code of conduct.

By participating in this project, you agree to abide by its [Code of Conduct](./CODE_OF_CONDUCT.md) at all times.

## Licensing

Copyright (c) 2023 Deutsche Telekom AG.

Licensed under the **Apache License, Version 2.0** (the "License"); you may not use this file except in compliance with the License.

You may obtain a copy of the License at https://www.apache.org/licenses/LICENSE-2.0.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the [LICENSE](./LICENSE) for the specific language governing permissions and limitations under the License.
