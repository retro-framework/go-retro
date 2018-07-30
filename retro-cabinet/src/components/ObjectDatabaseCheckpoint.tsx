import * as React from 'react';

class ObjectDatabaseCheckpoint extends React.Component<any, any> {
    constructor(props: any) {
        super(props);
    }
    public render() {
        if (this.props.commandDesc && this.props.hash) {
            return (
                <div className="obdv__checkpoint" onClick={this.props.selectCheckpointFn}>
                    <span className="odbv__checkpointHash" title={this.props.hash} >{this.props.hash.substr(7, 8)}</span>
                    <span className="odbv__checkpointSubject">{window.atob(this.props.commandDesc)}</span>
                </div>
            )
        } else {
            return (<div>Cannot Render Checkpoint</div>);
        }
    }
}

export default ObjectDatabaseCheckpoint;
