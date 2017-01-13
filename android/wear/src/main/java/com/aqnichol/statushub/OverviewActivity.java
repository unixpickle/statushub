package com.aqnichol.statushub;

import android.app.Activity;
import android.os.Bundle;
import android.support.wearable.view.WatchViewStub;
import android.view.View;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;

public class OverviewActivity extends Activity implements WatchViewStub.OnLayoutInflatedListener {

    private Button refreshButton;
    private LinearLayout listView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_overview);
        final WatchViewStub stub = (WatchViewStub) findViewById(R.id.watch_view_stub);
        stub.setOnLayoutInflatedListener(this);
    }

    @Override
    public void onLayoutInflated(WatchViewStub stub) {
        refreshButton = (Button)findViewById(R.id.refresh);
        refreshButton.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                refresh();
            }
        });

        listView = (LinearLayout)findViewById(R.id.overview_list);
        listView.addView(new LogEntry(this, "Service", "Message here."));
        listView.addView(new LogEntry(this, "Service", "The quick brown fox jumps over the lazy yet long log message."));
    }

    private void refresh() {
        // TODO: this.
        refreshButton.setEnabled(false);
    }
}
