import * as React from 'react';

// import ObjectDatabaseAffix from './ObjectDatabaseAffix';

// interface IObjectDatabaseCheckpointProps {
//     refHash: string;
// }

/* tslint:disable:no-console */

class ObjectDatabaseCheckpoint extends React.Component<{}, any> {
    constructor(props: any) {
        super(props);
        this.state = {
            checkpoint: {
                affixHash: "",
                commandDesc: "",
                hash: props.refHash,
                parentHashes: [],
            },
        };
    }
    public componentDidMount() {
        // fetch(`/obj/${this.props.refHash}`)
        //     .then((response) => {
        //         return response.json()
        //     })
        //     .then((cp) => {
        //         cp.hash = this.props.refHash; // needed because cp from server has no hash prop
        //         this.setState({ checkpoint: cp });
        //     })
        //     .catch((err) => console.error)
    }
    public render() {
        // console.log("in odcp render");
        // let pcp = Object.keys(this.state.checkpoint.parentHashes).map((_, parentHash) =>
        //     <ObjectDatabaseCheckpoint key={parentHash} refHash="{parentHash}" />
        // );
        // pcp = pcp
        return (
            <div className="obdv__checkpoint">
                {/* <span className="odbv__checkpointHash">{this.state.checkpoint.hash.substring(7, 15)}</span> */}
                <span className="odbv__refIndicator" title="I am not wired to anything">refs/heads/notwired</span>
                {/* <span className="odbv__checkpointSubject">{window.atob(this.state.checkpoint.commandDesc)}</span> */}
                <span className="odbv__datestamp" title="I am not wired to anything">??h ago</span>
                <span className="odbv__authorName" title="I am not wired to anything">Max Mustermann</span>
                {/* <ObjectDatabaseAffix refHash={this.state.checkpoint.affixHash} /> */}
            </div>
        )
    }
}

export default ObjectDatabaseCheckpoint;
