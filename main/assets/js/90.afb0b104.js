(window.webpackJsonp=window.webpackJsonp||[]).push([[90],{654:function(e,t,a){"use strict";a.r(t);var i=a(1),s=Object(i.a)({},(function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[a("h1",{attrs:{id:"adr-006-ics-02-client-refactor"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#adr-006-ics-02-client-refactor"}},[e._v("#")]),e._v(" ADR 006: ICS-02 client refactor")]),e._v(" "),a("h2",{attrs:{id:"changelog"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#changelog"}},[e._v("#")]),e._v(" Changelog")]),e._v(" "),a("ul",[a("li",[e._v("2022-08-01: Initial Draft")])]),e._v(" "),a("h2",{attrs:{id:"status"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#status"}},[e._v("#")]),e._v(" Status")]),e._v(" "),a("p",[e._v("Accepted and applied in v7 of ibc-go")]),e._v(" "),a("h2",{attrs:{id:"context"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#context"}},[e._v("#")]),e._v(" Context")]),e._v(" "),a("p",[e._v("During the initial development of the 02-client submodule, each light client supported (06-solomachine, 07-tendermint, 09-localhost) was referenced through hardcoding.\nHere is an example of the "),a("a",{attrs:{href:"https://github.com/cosmos/cosmos-sdk/commit/b93300288e3a04faef9c0774b75c13b24450ba1c#diff-c5f6b956947375f28d611f18d0e670cf28f8f305300a89c5a9b239b0eeec5064R83",target:"_blank",rel:"noopener noreferrer"}},[e._v("code"),a("OutboundLink")],1),e._v(" that existed in the 02-client submodule:")]),e._v(" "),a("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"ZnVuYyAoayBLZWVwZXIpIFVwZGF0ZUNsaWVudChjdHggc2RrLkNvbnRleHQsIGNsaWVudElEIHN0cmluZywgaGVhZGVyIGV4cG9ydGVkLkhlYWRlcikgKGV4cG9ydGVkLkNsaWVudFN0YXRlLCBlcnJvcikgewogIC4uLgoKICBzd2l0Y2ggY2xpZW50VHlwZSB7CiAgY2FzZSBleHBvcnRlZC5UZW5kZXJtaW50OgogICAgY2xpZW50U3RhdGUsIGNvbnNlbnN1c1N0YXRlLCBlcnIgPSB0ZW5kZXJtaW50LkNoZWNrVmFsaWRpdHlBbmRVcGRhdGVTdGF0ZSgKICAgIGNsaWVudFN0YXRlLCBoZWFkZXIsIGN0eC5CbG9ja1RpbWUoKSwKICAgICkKICBjYXNlIGV4cG9ydGVkLkxvY2FsaG9zdDoKICAgIC8vIG92ZXJyaWRlIGNsaWVudCBzdGF0ZSBhbmQgdXBkYXRlIHRoZSBibG9jayBoZWlnaHQKICAgIGNsaWVudFN0YXRlID0gbG9jYWxob3N0dHlwZXMuTmV3Q2xpZW50U3RhdGUoCiAgICBjdHguQ2hhaW5JRCgpLCAvLyB1c2UgdGhlIGNoYWluIElEIGZyb20gY29udGV4dCBzaW5jZSB0aGUgY2xpZW50IGlzIGZyb20gdGhlIHJ1bm5pbmcgY2hhaW4gKGkuZSBzZWxmKS4KICAgIGN0eC5CbG9ja0hlaWdodCgpLAogICAgKQogIGRlZmF1bHQ6CiAgICBlcnIgPSB0eXBlcy5FcnJJbnZhbGlkQ2xpZW50VHlwZQogIH0K"}}),e._v(" "),a("p",[e._v("To add additional light clients, code would need to be added directly to the 02-client submodule.\nEvidently, this would likely become problematic as IBC scaled to many chains using consensus mechanisms beyond the initial supported light clients.\nIssue "),a("a",{attrs:{href:"https://github.com/cosmos/cosmos-sdk/issues/6064",target:"_blank",rel:"noopener noreferrer"}},[e._v("#6064"),a("OutboundLink")],1),e._v(" on the SDK addressed this problem by creating a more modular 02-client submodule.\nThe 02-client submodule would now interact with each light client via an interface.\nWhile, this change was positive in development, increasing the flexibility and adoptability of IBC, it also opened the door to new problems.")]),e._v(" "),a("p",[e._v("The difficulty of generalizing light clients became apparent once changes to those light clients were required.\nEach light client represents a different consensus algorithm which may contain a host of complexity and nuances.\nHere are some examples of issues which arose for light clients that are not applicable to all the light clients supported (06-solomachine, 07-tendermint, 09-localhost):")]),e._v(" "),a("h3",{attrs:{id:"tendermint-non-zero-height-upgrades"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#tendermint-non-zero-height-upgrades"}},[e._v("#")]),e._v(" Tendermint non-zero height upgrades")]),e._v(" "),a("p",[e._v("Before the launch of IBC, it was determined that the golang implementation of "),a("a",{attrs:{href:"https://github.com/tendermint/tendermint",target:"_blank",rel:"noopener noreferrer"}},[e._v("tendermint"),a("OutboundLink")],1),e._v(" would not be capable of supporting non-zero height upgrades.\nThis implies that any upgrade would require changing of the chain ID and resetting the height to 0.\nA chain is uniquely identified by its chain-id and validator set.\nTwo different chain ID's can be viewed as different chains and thus a normal update produced by a validator set cannot change the chain ID.\nTo work around the lack of support for non-zero height upgrades, an abstract height type was created along with an upgrade mechanism.\nThis type would indicate the revision number (the number of times the chain ID has been changed) and revision height (the current height of the blockchain).")]),e._v(" "),a("p",[e._v("Refs:")]),e._v(" "),a("ul",[a("li",[e._v("Issue "),a("a",{attrs:{href:"https://github.com/cosmos/ibc/issues/439",target:"_blank",rel:"noopener noreferrer"}},[e._v("#439"),a("OutboundLink")],1),e._v(" on IBC specification repository.")]),e._v(" "),a("li",[e._v("Specification changes in "),a("a",{attrs:{href:"https://github.com/cosmos/ibc/pull/447",target:"_blank",rel:"noopener noreferrer"}},[e._v("#447"),a("OutboundLink")],1)]),e._v(" "),a("li",[e._v("Implementation changes for the abstract height type, "),a("a",{attrs:{href:"https://github.com/cosmos/cosmos-sdk/pull/7211",target:"_blank",rel:"noopener noreferrer"}},[e._v("SDK#7211"),a("OutboundLink")],1)])]),e._v(" "),a("h3",{attrs:{id:"tendermint-requires-misbehaviour-detection-during-updates"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#tendermint-requires-misbehaviour-detection-during-updates"}},[e._v("#")]),e._v(" Tendermint requires misbehaviour detection during updates")]),e._v(" "),a("p",[e._v("The initial release of the IBC module and the 07-tendermint light client implementation did not support misbehaviour detection during update nor did it prevent overwriting of previous updates.\nDespite the fact that we designed the "),a("code",[e._v("ClientState")]),e._v(" interface and developed the 07-tendermint client, we failed to detect even a duplicate update that constituted misbehaviour and thus should freeze the client.\nThis was fixed in PR "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/pull/141",target:"_blank",rel:"noopener noreferrer"}},[e._v("#141"),a("OutboundLink")],1),e._v(" which required light client implementations to be aware that they must handle duplicate updates and misbehaviour detection.\nMisbehaviour detection during updates is not applicable to the solomachine nor localhost.\nIt is also not obvious that "),a("code",[e._v("CheckHeaderAndUpdateState")]),e._v(" should be performing this functionality.")]),e._v(" "),a("h3",{attrs:{id:"localhost-requires-access-to-the-entire-client-store"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#localhost-requires-access-to-the-entire-client-store"}},[e._v("#")]),e._v(" Localhost requires access to the entire client store")]),e._v(" "),a("p",[e._v("The localhost has been broken since the initial version of the IBC module.\nThe localhost tried to be developed underneath the 02-client interfaces without special exception, but this proved to be impossible.\nThe issues were outlined in "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/issues/27",target:"_blank",rel:"noopener noreferrer"}},[e._v("#27"),a("OutboundLink")],1),e._v(" and further discussed in the attempted ADR in "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/pull/75",target:"_blank",rel:"noopener noreferrer"}},[e._v("#75"),a("OutboundLink")],1),e._v(".\nUnlike all other clients, the localhost requires access to the entire IBC store and not just the prefixed client store.")]),e._v(" "),a("h3",{attrs:{id:"solomachine-doesn-t-set-consensus-states"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#solomachine-doesn-t-set-consensus-states"}},[e._v("#")]),e._v(" Solomachine doesn't set consensus states")]),e._v(" "),a("p",[e._v("The 06-solomachine does not set the consensus states within the prefixed client store.\nIt has a single consensus state that is stored within the client state.\nThis causes setting of the consensus state at the 02-client level to use unnecessary storage.\nIt also causes timeouts to fail with solo machines.\nPreviously, the timeout logic within IBC would obtain the consensus state at the height a timeout is being proved.\nThis is problematic for the solo machine as no consensus state is set.\nSee issue "),a("a",{attrs:{href:"https://github.com/cosmos/ibc/issues/562",target:"_blank",rel:"noopener noreferrer"}},[e._v("#562"),a("OutboundLink")],1),e._v(" on the IBC specification repo.")]),e._v(" "),a("h3",{attrs:{id:"new-clients-may-want-to-do-batch-updates"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#new-clients-may-want-to-do-batch-updates"}},[e._v("#")]),e._v(" New clients may want to do batch updates")]),e._v(" "),a("p",[e._v("New light clients may not function in a similar fashion to 06-solomachine and 07-tendermint.\nThey may require setting many consensus states in a single update.\nAs @seunlanlege "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/issues/284#issuecomment-1005583679",target:"_blank",rel:"noopener noreferrer"}},[e._v("states"),a("OutboundLink")],1),e._v(":")]),e._v(" "),a("blockquote",[a("p",[e._v("I'm in support of these changes for 2 reasons:")]),e._v(" "),a("ul",[a("li",[a("p",[e._v("This would allow light clients handle batch header updates in CheckHeaderAndUpdateState, for the special case of 11-beefy proving the finality for a batch of headers is much more space and time efficient than the space/time complexity of proving each individual headers in that batch, combined.")])]),e._v(" "),a("li",[a("p",[e._v("This also allows for a single light client instance of 11-beefy be used to prove finality for every parachain connected to the relay chain (Polkadot/Kusama). We achieve this by setting the appropriate ConsensusState for individual parachain headers in CheckHeaderAndUpdateState")])])])]),e._v(" "),a("h2",{attrs:{id:"decision"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#decision"}},[e._v("#")]),e._v(" Decision")]),e._v(" "),a("h3",{attrs:{id:"require-light-clients-to-set-client-and-consensus-states"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#require-light-clients-to-set-client-and-consensus-states"}},[e._v("#")]),e._v(" Require light clients to set client and consensus states")]),e._v(" "),a("p",[e._v("The IBC specification states:")]),e._v(" "),a("blockquote",[a("p",[e._v("If the provided header was valid, the client MUST also mutate internal state to store now-finalised consensus roots and update any necessary signature authority tracking (e.g. changes to the validator set) for future calls to the validity predicate.")])]),e._v(" "),a("p",[e._v('The initial version of the IBC go SDK based module did not fulfill this requirement.\nInstead, the 02-client submodule required each light client to return the client and consensus state which should be updated in the client prefixed store.\nThis decision lead to the issues "Solomachine doesn\'t set consensus states" and "New clients may want to do batch updates".')]),e._v(" "),a("p",[e._v("Each light client should be required to set its own client and consensus states on any update necessary.\nThe go implementation should be changed to match the specification requirements.\nThis will allow more flexibility for light clients to manage their own internal storage and do batch updates.")]),e._v(" "),a("h3",{attrs:{id:"merge-header-misbehaviour-interface-and-rename-to-clientmessage"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#merge-header-misbehaviour-interface-and-rename-to-clientmessage"}},[e._v("#")]),e._v(" Merge "),a("code",[e._v("Header")]),e._v("/"),a("code",[e._v("Misbehaviour")]),e._v(" interface and rename to "),a("code",[e._v("ClientMessage")])]),e._v(" "),a("p",[e._v("Remove "),a("code",[e._v("GetHeight()")]),e._v(" from the header interface (as light clients now set the client/consensus states).\nThis results in the "),a("code",[e._v("Header")]),e._v("/"),a("code",[e._v("Misbehaviour")]),e._v(" interfaces being the same.\nTo reduce complexity of the codebase, the "),a("code",[e._v("Header")]),e._v("/"),a("code",[e._v("Misbehaviour")]),e._v(" interfaces should be merged into "),a("code",[e._v("ClientMessage")]),e._v(".\n"),a("code",[e._v("ClientMessage")]),e._v(" will provide the client with some authenticated information which may result in regular updates, misbehaviour detection, batch updates, or other custom functionality a light client requires.")]),e._v(" "),a("h3",{attrs:{id:"split-checkheaderandupdatestate-into-4-functions"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#split-checkheaderandupdatestate-into-4-functions"}},[e._v("#")]),e._v(" Split "),a("code",[e._v("CheckHeaderAndUpdateState")]),e._v(" into 4 functions")]),e._v(" "),a("p",[e._v("See "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/issues/668",target:"_blank",rel:"noopener noreferrer"}},[e._v("#668"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("p",[e._v("Split "),a("code",[e._v("CheckHeaderAndUpdateState")]),e._v(" into 4 functions:")]),e._v(" "),a("ul",[a("li",[a("code",[e._v("VerifyClientMessage")])]),e._v(" "),a("li",[a("code",[e._v("CheckForMisbehaviour")])]),e._v(" "),a("li",[a("code",[e._v("UpdateStateOnMisbehaviour")])]),e._v(" "),a("li",[a("code",[e._v("UpdateState")])])]),e._v(" "),a("p",[a("code",[e._v("VerifyClientMessage")]),e._v(" checks the that the structure of a "),a("code",[e._v("ClientMessage")]),e._v(" is correct and that all authentication data provided is valid.")]),e._v(" "),a("p",[a("code",[e._v("CheckForMisbehaviour")]),e._v(" checks to see if a "),a("code",[e._v("ClientMessage")]),e._v(" is evidence of misbehaviour.")]),e._v(" "),a("p",[a("code",[e._v("UpdateStateOnMisbehaviour")]),e._v(" freezes the client and updates its state accordingly.")]),e._v(" "),a("p",[a("code",[e._v("UpdateState")]),e._v(" performs a regular update or a no-op on duplicate updates.")]),e._v(" "),a("p",[e._v("The code roughly looks like:")]),e._v(" "),a("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"ZnVuYyAoayBLZWVwZXIpIFVwZGF0ZUNsaWVudChjdHggc2RrLkNvbnRleHQsIGNsaWVudElEIHN0cmluZywgaGVhZGVyIGV4cG9ydGVkLkhlYWRlcikgZXJyb3IgewogIC4uLgoKICBpZiBlcnIgOj0gY2xpZW50U3RhdGUuVmVyaWZ5Q2xpZW50TWVzc2FnZShjbGllbnRNZXNzYWdlKTsgZXJyICE9IG5pbCB7CiAgICByZXR1cm4gZXJyCiAgfQogIAogIGZvdW5kTWlzYmVoYXZpb3VyIDo9IGNsaWVudFN0YXRlLkNoZWNrRm9yTWlzYmVoYXZpb3VyKGNsaWVudE1lc3NhZ2UpCiAgaWYgZm91bmRNaXNiZWhhdmlvdXIgewogICAgY2xpZW50U3RhdGUuVXBkYXRlU3RhdGVPbk1pc2JlaGF2aW91cihoZWFkZXIpCiAgICAvLyBlbWl0IG1pc2JlaGF2aW91ciBldmVudAogICAgcmV0dXJuIAogIH0KICAKICBjbGllbnRTdGF0ZS5VcGRhdGVTdGF0ZShjbGllbnRNZXNzYWdlKSAvLyBleHBlY3RzIG5vLW9wIG9uIGR1cGxpY2F0ZSBoZWFkZXIKICAvLyBlbWl0IHVwZGF0ZSBldmVudAogIHJldHVybgp9Cg=="}}),e._v(" "),a("h3",{attrs:{id:"add-gettimestampatheight-to-the-client-state-interface"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#add-gettimestampatheight-to-the-client-state-interface"}},[e._v("#")]),e._v(" Add "),a("code",[e._v("GetTimestampAtHeight")]),e._v(" to the client state interface")]),e._v(" "),a("p",[e._v("By adding "),a("code",[e._v("GetTimestampAtHeight")]),e._v(" to the ClientState interface, we allow light clients which do non-traditional consensus state/timestamp storage to process timeouts correctly.\nThis fixes the issues outlined for the solo machine client.")]),e._v(" "),a("h3",{attrs:{id:"add-generic-verification-functions"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#add-generic-verification-functions"}},[e._v("#")]),e._v(" Add generic verification functions")]),e._v(" "),a("p",[e._v("As the complexity and the functionality grows, new verification functions will be required for additional paths.\nThis was explained in "),a("a",{attrs:{href:"https://github.com/cosmos/ibc/issues/684",target:"_blank",rel:"noopener noreferrer"}},[e._v("#684"),a("OutboundLink")],1),e._v(" on the specification repo.\nThese generic verification functions would be immediately useful for the new paths added in connection/channel upgradability as well as for custom paths defined by IBC applications such as Interchain Queries.\nThe old verification functions ("),a("code",[e._v("VerifyClientState")]),e._v(", "),a("code",[e._v("VerifyConnection")]),e._v(", etc) should be removed in favor of the generic verification functions.")]),e._v(" "),a("h2",{attrs:{id:"consequences"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#consequences"}},[e._v("#")]),e._v(" Consequences")]),e._v(" "),a("h3",{attrs:{id:"positive"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#positive"}},[e._v("#")]),e._v(" Positive")]),e._v(" "),a("ul",[a("li",[e._v("Flexibility for light client implementations")]),e._v(" "),a("li",[e._v("Well defined interfaces and their required functionality")]),e._v(" "),a("li",[e._v("Generic verification functions")]),e._v(" "),a("li",[e._v("Applies changes necessary for future client/connection/channel upgrabability features")]),e._v(" "),a("li",[e._v("Timeout processing for solo machines")]),e._v(" "),a("li",[e._v("Reduced code complexity")])]),e._v(" "),a("h3",{attrs:{id:"negative"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#negative"}},[e._v("#")]),e._v(" Negative")]),e._v(" "),a("ul",[a("li",[e._v("The refactor touches on sensitive areas of the ibc-go codebase")]),e._v(" "),a("li",[e._v("Changing of established naming ("),a("code",[e._v("Header")]),e._v("/"),a("code",[e._v("Misbehaviour")]),e._v(" to "),a("code",[e._v("ClientMessage")]),e._v(")")])]),e._v(" "),a("h3",{attrs:{id:"neutral"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#neutral"}},[e._v("#")]),e._v(" Neutral")]),e._v(" "),a("p",[e._v("No notable consequences")]),e._v(" "),a("h2",{attrs:{id:"references"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#references"}},[e._v("#")]),e._v(" References")]),e._v(" "),a("p",[e._v("Issues:")]),e._v(" "),a("ul",[a("li",[a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/issues/284",target:"_blank",rel:"noopener noreferrer"}},[e._v("#284"),a("OutboundLink")],1)])]),e._v(" "),a("p",[e._v("PRs:")]),e._v(" "),a("ul",[a("li",[a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/pull/1871",target:"_blank",rel:"noopener noreferrer"}},[e._v("#1871"),a("OutboundLink")],1)])])],1)}),[],!1,null,null,null);t.default=s.exports}}]);