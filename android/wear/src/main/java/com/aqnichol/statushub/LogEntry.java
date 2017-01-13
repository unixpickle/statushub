package com.aqnichol.statushub;

import android.content.Context;
import android.graphics.Color;
import android.graphics.Typeface;
import android.util.Log;
import android.widget.LinearLayout;
import android.widget.TextView;

/**
 * A LogEntry shows a single log record.
 */
public class LogEntry extends LinearLayout {

    private TextView serviceName;
    private TextView message;

    /**
     * Top margin in dp.
     */
    private static final float MARGIN_TOP = 7;

    public LogEntry(Context context, String service, String msg) {
        super(context);
        init();
        serviceName.setText(service);
        message.setText(msg);
    }

    private void init() {
        LinearLayout.LayoutParams layout = new LinearLayout.LayoutParams(
                LayoutParams.MATCH_PARENT,
                LayoutParams.WRAP_CONTENT
        );
        float density = getContext().getResources().getDisplayMetrics().density;
        int topMargin = Math.round(density * MARGIN_TOP);
        layout.setMargins(0, topMargin, 0, 0);
        this.setLayoutParams(layout);
        this.setOrientation(LinearLayout.VERTICAL);

        serviceName = new TextView(this.getContext());
        serviceName.setLayoutParams(new LinearLayout.LayoutParams(
                LayoutParams.MATCH_PARENT,
                LayoutParams.WRAP_CONTENT
        ));
        serviceName.setTypeface(Typeface.DEFAULT_BOLD);
        serviceName.setTextColor(Color.WHITE);
        this.addView(serviceName);

        message = new TextView(this.getContext());
        message.setLayoutParams(new LinearLayout.LayoutParams(
                LayoutParams.MATCH_PARENT,
                LayoutParams.WRAP_CONTENT
        ));
        message.setTextColor(Color.WHITE);
        this.addView(message);
    }

}
