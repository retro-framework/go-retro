import * as React from 'react';

class ObjectDatabaseCheckpoint extends React.Component<any, any> {
    constructor(props: any) {
        super(props);
    }
    public render() {
        return (
            <div className="obdv__checkpoint">
                <span className="odbv__checkpointHash" title={this.props.hash} >{this.props.hash.substr(7, 8)}</span>
                <span className="odbv__checkpointSubject">{window.atob(this.props.commandDesc)}</span>
            </div>
        )
    }
}

export default ObjectDatabaseCheckpoint;
