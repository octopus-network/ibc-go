(window.webpackJsonp=window.webpackJsonp||[]).push([[133],{694:function(e,t,a){"use strict";a.r(t);var i=a(1),n=Object(i.a)({},(function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[a("h1",{attrs:{id:"handling-clientmessages-updates-and-misbehaviour"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#handling-clientmessages-updates-and-misbehaviour"}},[e._v("#")]),e._v(" Handling "),a("code",[e._v("ClientMessage")]),e._v("s: updates and misbehaviour")]),e._v(" "),a("p",[e._v("As mentioned before in the documentation about "),a("RouterLink",{attrs:{to:"/ibc/light-clients/consensus-state.html"}},[e._v("implementing the "),a("code",[e._v("ConsensusState")]),e._v(" interface")]),e._v(", "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/core/exported/client.go#L147",target:"_blank",rel:"noopener noreferrer"}},[a("code",[e._v("ClientMessage")]),a("OutboundLink")],1),e._v(" is an interface used to update an IBC client. This update may be performed by:")],1),e._v(" "),a("ul",[a("li",[e._v("a single header")]),e._v(" "),a("li",[e._v("a batch of headers")]),e._v(" "),a("li",[e._v("evidence of misbehaviour,")]),e._v(" "),a("li",[e._v("or any type which when verified produces a change to the consensus state of the IBC client.")])]),e._v(" "),a("p",[e._v("This interface has been purposefully kept generic in order to give the maximum amount of flexibility to the light client implementer.")]),e._v(" "),a("h2",{attrs:{id:"implementing-the-clientmessage-interface"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#implementing-the-clientmessage-interface"}},[e._v("#")]),e._v(" Implementing the "),a("code",[e._v("ClientMessage")]),e._v(" interface")]),e._v(" "),a("p",[e._v("Find the "),a("code",[e._v("ClientMessage")]),e._v("interface in "),a("code",[e._v("modules/core/exported")]),e._v(":")]),e._v(" "),a("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"dHlwZSBDbGllbnRNZXNzYWdlIGludGVyZmFjZSB7CiAgcHJvdG8uTWVzc2FnZQoKICBDbGllbnRUeXBlKCkgc3RyaW5nCiAgVmFsaWRhdGVCYXNpYygpIGVycm9yCn0K"}}),e._v(" "),a("p",[e._v("The "),a("code",[e._v("ClientMessage")]),e._v(" will be passed to the client to be used in "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/core/02-client/keeper/client.go#L48",target:"_blank",rel:"noopener noreferrer"}},[a("code",[e._v("UpdateClient")]),a("OutboundLink")],1),e._v(", which retrieves the "),a("code",[e._v("ClientState")]),e._v(" by client ID (available in "),a("code",[e._v("MsgUpdateClient")]),e._v("). This "),a("code",[e._v("ClientState")]),e._v(" implements the "),a("RouterLink",{attrs:{to:"/ibc/light-clients/client-state.html"}},[a("code",[e._v("ClientState")]),e._v(" interface")]),e._v(" for its specific consenus type (e.g. Tendermint).")],1),e._v(" "),a("p",[a("code",[e._v("UpdateClient")]),e._v(" will then handle a number of cases including misbehaviour and/or updating the consensus state, utilizing the specific methods defined in the relevant "),a("code",[e._v("ClientState")]),e._v(".")]),e._v(" "),a("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"VmVyaWZ5Q2xpZW50TWVzc2FnZShjdHggc2RrLkNvbnRleHQsIGNkYyBjb2RlYy5CaW5hcnlDb2RlYywgY2xpZW50U3RvcmUgc2RrLktWU3RvcmUsIGNsaWVudE1zZyBDbGllbnRNZXNzYWdlKSBlcnJvcgpDaGVja0Zvck1pc2JlaGF2aW91cihjdHggc2RrLkNvbnRleHQsIGNkYyBjb2RlYy5CaW5hcnlDb2RlYywgY2xpZW50U3RvcmUgc2RrLktWU3RvcmUsIGNsaWVudE1zZyBDbGllbnRNZXNzYWdlKSBib29sClVwZGF0ZVN0YXRlT25NaXNiZWhhdmlvdXIoY3R4IHNkay5Db250ZXh0LCBjZGMgY29kZWMuQmluYXJ5Q29kZWMsIGNsaWVudFN0b3JlIHNkay5LVlN0b3JlLCBjbGllbnRNc2cgQ2xpZW50TWVzc2FnZSkKVXBkYXRlU3RhdGUoY3R4IHNkay5Db250ZXh0LCBjZGMgY29kZWMuQmluYXJ5Q29kZWMsIGNsaWVudFN0b3JlIHNkay5LVlN0b3JlLCBjbGllbnRNc2cgQ2xpZW50TWVzc2FnZSkgW11IZWlnaHQK"}}),e._v(" "),a("h2",{attrs:{id:"handling-updates-and-misbehaviour"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#handling-updates-and-misbehaviour"}},[e._v("#")]),e._v(" Handling updates and misbehaviour")]),e._v(" "),a("p",[e._v("The functions for handling updates to a light client and evidence of misbehaviour are all found in the "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/core/exported/client.go#L36",target:"_blank",rel:"noopener noreferrer"}},[a("code",[e._v("ClientState")]),a("OutboundLink")],1),e._v(" interface, and will be discussed below.")]),e._v(" "),a("blockquote",[a("p",[e._v("It is important to note that "),a("code",[e._v("Misbehaviour")]),e._v(" in this particular context is referring to misbehaviour on the chain level intended to fool the light client. This will be defined by each light client.")])]),e._v(" "),a("h2",{attrs:{id:"verifyclientmessage"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#verifyclientmessage"}},[e._v("#")]),e._v(" "),a("code",[e._v("VerifyClientMessage")])]),e._v(" "),a("p",[a("code",[e._v("VerifyClientMessage")]),e._v(" must verify a "),a("code",[e._v("ClientMessage")]),e._v(". A "),a("code",[e._v("ClientMessage")]),e._v(" could be a "),a("code",[e._v("Header")]),e._v(", "),a("code",[e._v("Misbehaviour")]),e._v(", or batch update. To understand how to implement a "),a("code",[e._v("ClientMessage")]),e._v(", please refer to the "),a("a",{attrs:{href:"#implementing-the-clientmessage-interface"}},[e._v("Implementing the "),a("code",[e._v("ClientMessage")]),e._v(" interface")]),e._v(" section.")]),e._v(" "),a("p",[e._v("It must handle each type of "),a("code",[e._v("ClientMessage")]),e._v(" appropriately. Calls to "),a("code",[e._v("CheckForMisbehaviour")]),e._v(", "),a("code",[e._v("UpdateState")]),e._v(", and "),a("code",[e._v("UpdateStateOnMisbehaviour")]),e._v(" will assume that the content of the "),a("code",[e._v("ClientMessage")]),e._v(" has been verified and can be trusted. An error should be returned if the "),a("code",[e._v("ClientMessage")]),e._v(" fails to verify.")]),e._v(" "),a("p",[e._v("For an example of a "),a("code",[e._v("VerifyClientMessage")]),e._v(" implementation, please check the "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/light-clients/07-tendermint/update.go#L20",target:"_blank",rel:"noopener noreferrer"}},[e._v("Tendermint light client"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("h2",{attrs:{id:"checkformisbehaviour"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#checkformisbehaviour"}},[e._v("#")]),e._v(" "),a("code",[e._v("CheckForMisbehaviour")])]),e._v(" "),a("p",[e._v("Checks for evidence of a misbehaviour in "),a("code",[e._v("Header")]),e._v(" or "),a("code",[e._v("Misbehaviour")]),e._v(" type. It assumes the "),a("code",[e._v("ClientMessage")]),e._v(" has already been verified.")]),e._v(" "),a("p",[e._v("For an example of a "),a("code",[e._v("CheckForMisbehaviour")]),e._v(" implementation, please check the "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/light-clients/07-tendermint/misbehaviour_handle.go#L19",target:"_blank",rel:"noopener noreferrer"}},[e._v("Tendermint light client"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("blockquote",[a("p",[e._v("The Tendermint light client "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/light-clients/07-tendermint/misbehaviour.go",target:"_blank",rel:"noopener noreferrer"}},[e._v("defines "),a("code",[e._v("Misbehaviour")]),a("OutboundLink")],1),e._v(" as two different types of situations: a situation where two conflicting "),a("code",[e._v("Header")]),e._v("s with the same height have been submitted to update a client's "),a("code",[e._v("ConsensusState")]),e._v(" within the same trusting period, or that the two conflicting "),a("code",[e._v("Header")]),e._v("s have been submitted at different heights but the consensus states are not in the correct monotonic time ordering (BFT time violation). More explicitly, updating to a new height must have a timestamp greater than the previous consensus state, or, if inserting a consensus at a past height, then time must be less than those heights which come after and greater than heights which come before.")])]),e._v(" "),a("h2",{attrs:{id:"updatestateonmisbehaviour"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#updatestateonmisbehaviour"}},[e._v("#")]),e._v(" "),a("code",[e._v("UpdateStateOnMisbehaviour")])]),e._v(" "),a("p",[a("code",[e._v("UpdateStateOnMisbehaviour")]),e._v(" should perform appropriate state changes on a client state given that misbehaviour has been detected and verified. This method should only be called when misbehaviour is detected, as it does not perform any misbehaviour checks. Notably, it should freeze the client so that calling the "),a("code",[e._v("Status")]),e._v(" function on the associated client state no longer returns "),a("code",[e._v("Active")]),e._v(".")]),e._v(" "),a("p",[e._v("For an example of a "),a("code",[e._v("UpdateStateOnMisbehaviour")]),e._v(" implementation, please check the "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/light-clients/07-tendermint/update.go#L199",target:"_blank",rel:"noopener noreferrer"}},[e._v("Tendermint light client"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("h2",{attrs:{id:"updatestate"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#updatestate"}},[e._v("#")]),e._v(" "),a("code",[e._v("UpdateState")])]),e._v(" "),a("p",[a("code",[e._v("UpdateState")]),e._v(" updates and stores as necessary any associated information for an IBC client, such as the "),a("code",[e._v("ClientState")]),e._v(" and corresponding "),a("code",[e._v("ConsensusState")]),e._v(". It should perform a no-op on duplicate updates.")]),e._v(" "),a("p",[e._v("It assumes the "),a("code",[e._v("ClientMessage")]),e._v(" has already been verified.")]),e._v(" "),a("p",[e._v("For an example of a "),a("code",[e._v("UpdateState")]),e._v(" implementation, please check the "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/light-clients/07-tendermint/update.go#L131",target:"_blank",rel:"noopener noreferrer"}},[e._v("Tendermint light client"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("h2",{attrs:{id:"putting-it-all-together"}},[a("a",{staticClass:"header-anchor",attrs:{href:"#putting-it-all-together"}},[e._v("#")]),e._v(" Putting it all together")]),e._v(" "),a("p",[e._v("The "),a("code",[e._v("02-client")]),e._v(" "),a("code",[e._v("Keeper")]),e._v(" module in ibc-go offers a reference as to how these functions will be used to "),a("a",{attrs:{href:"https://github.com/cosmos/ibc-go/blob/v7.0.0/modules/core/02-client/keeper/client.go#L48",target:"_blank",rel:"noopener noreferrer"}},[e._v("update the client"),a("OutboundLink")],1),e._v(".")]),e._v(" "),a("tm-code-block",{staticClass:"codeblock",attrs:{language:"go",base64:"aWYgZXJyIDo9IGNsaWVudFN0YXRlLlZlcmlmeUNsaWVudE1lc3NhZ2UoY2xpZW50TWVzc2FnZSk7IGVyciAhPSBuaWwgewogIHJldHVybiBlcnIKfQoKZm91bmRNaXNiZWhhdmlvdXIgOj0gY2xpZW50U3RhdGUuQ2hlY2tGb3JNaXNiZWhhdmlvdXIoY2xpZW50TWVzc2FnZSkKaWYgZm91bmRNaXNiZWhhdmlvdXIgewogIGNsaWVudFN0YXRlLlVwZGF0ZVN0YXRlT25NaXNiZWhhdmlvdXIoY2xpZW50TWVzc2FnZSkKICAvLyBlbWl0IG1pc2JlaGF2aW91ciBldmVudAogIHJldHVybiAKfQoKY2xpZW50U3RhdGUuVXBkYXRlU3RhdGUoY2xpZW50TWVzc2FnZSkgLy8gZXhwZWN0cyBuby1vcCBvbiBkdXBsaWNhdGUgaGVhZGVyCi8vIGVtaXQgdXBkYXRlIGV2ZW50CnJldHVybgo="}})],1)}),[],!1,null,null,null);t.default=n.exports}}]);