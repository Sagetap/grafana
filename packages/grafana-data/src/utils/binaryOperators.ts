import { RegistryItem, Registry } from './Registry';

export enum BinaryOperationID {
  Add = '+',
  Subtract = '-',
  Divide = '/',
  DivideAndRound = '\u230A/\u2309',
  Multiply = '*',
  TimeBucket = '\u23b5',
}

export type BinaryOperation = (left: number, right: number) => number;

interface BinaryOperatorInfo extends RegistryItem {
  operation: BinaryOperation;
}

export const binaryOperators = new Registry<BinaryOperatorInfo>(() => {
  return [
    {
      id: BinaryOperationID.Add,
      name: 'Add',
      operation: (a: number, b: number) => a + b,
    },
    {
      id: BinaryOperationID.Subtract,
      name: 'Subtract',
      operation: (a: number, b: number) => a - b,
    },
    {
      id: BinaryOperationID.Multiply,
      name: 'Multiply',
      operation: (a: number, b: number) => a * b,
    },
    {
      id: BinaryOperationID.Divide,
      name: 'Divide',
      operation: (a: number, b: number) => a / b,
    },
    {
      id: BinaryOperationID.DivideAndRound,
      name: 'Divide and round',
      operation: (a: number, b: number) => Math.round(a / b),
    },
    {
      id: BinaryOperationID.TimeBucket,
      name: 'Time bucket',
      operation: (a: number, b: number) => Math.round(a / b) * b,
    },
  ];
});
