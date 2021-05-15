import React from 'react';
import {
  PanelContext,
  PanelContextRoot,
  GraphNG,
  GraphNGProps,
  BarValueVisibility,
  preparePlotFrame,
} from '@grafana/ui';
import { DataFrame, FieldType, TimeRange } from '@grafana/data';
import { preparePlotConfigBuilder } from './utils';
import { TimelineMode, TimelineValueAlignment } from './types';
import { XYFieldMatchers } from '@grafana/ui/src/components/GraphNG/types';

/**
 * @alpha
 */
export interface TimelineProps extends Omit<GraphNGProps, 'prepConfig' | 'propsToDiff' | 'renderLegend'> {
  mode: TimelineMode;
  rowHeight: number;
  showValue: BarValueVisibility;
  alignValue: TimelineValueAlignment;
  colWidth?: number;
}

const propsToDiff = ['mode', 'rowHeight', 'colWidth', 'showValue', 'alignValue'];

const _preparePlotFrame = (frames: DataFrame[], dimFields: XYFieldMatchers) => {
  frames.forEach((frame) => {
    frame.fields.forEach((field) => {
      field.config.custom = field.config.custom || {};
      field.config.custom.spanNulls = -1;
    });
  });

  return preparePlotFrame(frames, dimFields);
};

export class TimelineChart extends React.Component<TimelineProps> {
  static contextType = PanelContextRoot;
  panelContext: PanelContext = {} as PanelContext;

  prepConfig = (alignedFrame: DataFrame, getTimeRange: () => TimeRange) => {
    this.panelContext = this.context as PanelContext;
    const { eventBus } = this.panelContext;

    return preparePlotConfigBuilder({
      frame: alignedFrame,
      getTimeRange,
      eventBus,
      ...this.props,
    });
  };

  renderLegend = () => null;

  render() {
    return (
      <GraphNG
        {...this.props}
        fields={{
          x: (f) => f.type === FieldType.time,
          y: (f) => f.type === FieldType.number || f.type === FieldType.boolean || f.type === FieldType.string,
        }}
        preparePlotFrame={_preparePlotFrame as any}
        prepConfig={this.prepConfig}
        propsToDiff={propsToDiff}
        renderLegend={this.renderLegend}
      />
    );
  }
}
