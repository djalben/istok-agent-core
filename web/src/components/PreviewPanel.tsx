interface PreviewPanelProps {
  active: boolean;
}

const PreviewPanel = ({ active }: PreviewPanelProps) => {
  return (
    <div className="h-full glass border-l border-border/50 flex flex-col animate-fade-in-up">
      <div className="h-10 border-b border-border/50 flex items-center px-4">
        <span className="text-xs text-muted-foreground uppercase tracking-widest font-medium">
          Предпросмотр
        </span>
      </div>
      <div className="flex-1 flex items-center justify-center">
        {active ? (
          <iframe
            title="preview"
            className="w-full h-full border-0"
            srcDoc="<html><body style='background:#0f0f12;color:#E1E1E6;font-family:Inter,sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;margin:0;'><div style='text-align:center'><p style='font-size:14px;opacity:0.6'>Генерация завершена</p><p style='font-size:12px;opacity:0.3;margin-top:8px'>Здесь будет ваше приложение</p></div></body></html>"
          />
        ) : (
          <div className="text-center px-6">
            <div className="text-muted-foreground text-sm">
              Результат генерации появится здесь
            </div>
            <div className="text-muted-foreground/40 text-xs mt-2">
              Введите описание и нажмите «Запустить генерацию»
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default PreviewPanel;
