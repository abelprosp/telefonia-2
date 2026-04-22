import { Skeleton } from '@/components/ui/skeleton';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow
} from '@/components/ui/table';

export type ListPageSkeletonColumn = {
  header: string;
  headClassName?: string;
  /** `text` = célula genérica; `actionsPair` = ícone abrir + ícone excluir; `actionsLink` = só ícone abrir */
  cell: 'text' | 'actionsPair' | 'actionsLink';
};

const TEXT_MAX_WIDTHS = [
  'max-w-[220px]',
  'max-w-[200px]',
  'max-w-[180px]',
  'max-w-[160px]',
  'max-w-[140px]'
];

type ListPageSkeletonProps = {
  pageSize: number;
  columns: ListPageSkeletonColumn[];
};

function SkeletonDataCell({
  variant,
  textColIndex
}: {
  variant: ListPageSkeletonColumn['cell'];
  textColIndex: number;
}) {
  if (variant === 'actionsPair') {
    return (
      <TableCell className="text-right">
        <div className="flex justify-end gap-2">
          <Skeleton className="size-7 shrink-0 rounded-md" />
          <Skeleton className="size-8 shrink-0 rounded-md" />
        </div>
      </TableCell>
    );
  }

  if (variant === 'actionsLink') {
    return (
      <TableCell className="text-right">
        <div className="flex justify-end gap-2">
          <Skeleton className="size-7 shrink-0 rounded-md" />
        </div>
      </TableCell>
    );
  }

  const w = TEXT_MAX_WIDTHS[textColIndex % TEXT_MAX_WIDTHS.length];
  return (
    <TableCell>
      <Skeleton className={`h-4 w-full ${w}`} />
    </TableCell>
  );
}

export function ListPageSkeleton({ pageSize, columns }: ListPageSkeletonProps) {
  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <Skeleton className="h-7 w-40 sm:w-48" />
        <Skeleton className="h-4 w-full max-w-md" />
      </div>
      <div className="overflow-hidden rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {columns.map((col) => (
                <TableHead key={col.header} className={col.headClassName}>
                  {col.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.from({ length: pageSize }).map((_, rowIndex) => {
              let textIdx = 0;
              return (
                <TableRow key={rowIndex}>
                  {columns.map((col, colIndex) => {
                    const node = (
                      <SkeletonDataCell
                        key={`${rowIndex}-${colIndex}`}
                        variant={col.cell}
                        textColIndex={textIdx}
                      />
                    );
                    if (col.cell === 'text') {
                      textIdx += 1;
                    }
                    return node;
                  })}
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </div>
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <Skeleton className="h-4 w-56" />
        <Skeleton className="h-8 w-full max-w-xs sm:ml-auto" />
      </div>
    </div>
  );
}
