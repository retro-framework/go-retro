import * as React from 'react';
import './RefSelector.css';

interface IRefSelectorProps {
  onChange: (newRef: string) => void
}

/* tslint:disable:no-console */

class RefSelector extends React.Component<IRefSelectorProps, { refs: {} }> {
  constructor(props: any) {
    super(props);
    this.state = {
      refs: {"Loading...": undefined}
    };
  }
  public componentDidMount() {
    fetch('/ref/')
    .then((response) => {
      return response.text().then((text) => {
        return text ? JSON.parse(text) : {}
      })
    })
    .then((refs) => {
      this.setState({refs});
      this.update();
    })
  }
  public update = () => {
    this.props.onChange(this.state.refs[Object.keys(this.state.refs)[Object.keys(this.state.refs).length - 1]]);
  }
  public render() {
    const options = Object.keys(this.state.refs).map((name, _) => 
      <option key={name} value={this.state.refs[name]}>{name}</option>
    );
    return (
      <select className="RefSelector" value={this.state.refs[Object.keys(this.state.refs)[Object.keys(this.state.refs).length - 1]]} onChange={this.update}>
        {options}
      </select>
    );
  }
}

export default RefSelector;
