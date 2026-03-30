import React from "react";
import { cn } from "@/lib/utils";

interface SettingTableColumn<T> {
  key: string;
  header: string;
  className?: string;
  render?: (value: any, row: T) => React.ReactNode;
}

interface SettingTableProps<T> {
  columns: SettingTableColumn<T>[];
  data: T[];
  emptyMessage?: string;
  className?: string;
  getRowKey?: (row: T, index: number) => string;
}

const SettingTable = <T extends Record<string, any>>({
  columns,
  data,
  emptyMessage = "No data available",
  className,
  getRowKey,
}: SettingTableProps<T>) => {
  return (
    <div className={cn("w-full overflow-x-auto border border-border rounded-xl bg-background", className)}>
      <table className="w-full text-left border-collapse">
        <thead>
          <tr className="border-b border-border bg-muted/30 text-xs font-semibold uppercase tracking-wider text-muted-foreground/80">
            {columns.map((column) => (
              <th key={column.key} className={cn("px-4 py-2.5", column.className)}>
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="divide-y divide-border">
          {data.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="px-4 py-8 text-center text-sm text-muted-foreground">
                {emptyMessage}
              </td>
            </tr>
          ) : (
            data.map((row, rowIndex) => {
              const rowKey = getRowKey ? getRowKey(row, rowIndex) : rowIndex.toString();
              return (
                <tr key={rowKey} className="hover:bg-muted/30 transition-colors">
                  {columns.map((column) => {
                    const value = row[column.key];
                    const content = column.render ? column.render(value, row) : (value as React.ReactNode);
                    return (
                      <td key={column.key} className={cn("px-4 py-3 text-sm text-foreground align-middle", column.className)}>
                        {content}
                      </td>
                    );
                  })}
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
};

export default SettingTable;
