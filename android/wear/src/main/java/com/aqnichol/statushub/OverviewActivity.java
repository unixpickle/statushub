package com.aqnichol.statushub;

import android.app.Activity;
import android.os.Bundle;
import android.support.wearable.view.WatchViewStub;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.LinearLayout;
import android.widget.TextView;

import com.aqnichol.watchcomm.CommException;
import com.aqnichol.watchcomm.LogEntries;
import com.aqnichol.watchcomm.Sender;
import com.aqnichol.watchcomm.Receiver;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.Wearable;

import org.json.JSONException;

import java.nio.charset.StandardCharsets;

public class OverviewActivity extends Activity implements WatchViewStub.OnLayoutInflatedListener {

    private final static int REFRESH_TIMEOUT = 10000;

    private Button refreshButton;
    private LinearLayout listView;
    private TextView errView;
    private Receiver receiver = new Receiver(this);

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_overview);
        final WatchViewStub stub = (WatchViewStub) findViewById(R.id.watch_view_stub);
        stub.setOnLayoutInflatedListener(this);
        receiver.connect();
    }

    @Override
    protected void onRestart() {
        super.onRestart();
        receiver.connect();
    }

    @Override
    protected void onStop() {
        super.onStop();
        receiver.disconnect();
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
        errView = (TextView)findViewById(R.id.errmsg);
        listView = (LinearLayout)findViewById(R.id.overview_list);
        listView.addView(new LogEntry(this, "No messages", "Refresh to see them."));
    }

    private void refresh() {
        refreshButton.setEnabled(false);
        new Thread(new Runnable() {
            @Override
            public void run() {
                Sender s = new Sender(getApplicationContext());
                try {
                    receiver.clearQueue();
                    s.connect();
                    s.sendMessage("/refresh", null);
                    MessageEvent evt = receiver.poll(REFRESH_TIMEOUT);
                    if (evt == null) {
                        displayErrorMessage("connection timeout");
                    } else if (evt.getPath().equals("/error")) {
                        String msg = new String(evt.getData(), StandardCharsets.UTF_8);
                        displayErrorMessage(msg);
                    } else {
                        displayList(evt);
                    }
                } catch (CommException e) {
                    displayErrorMessage(e.getMessage());
                } finally {
                    s.disconnect();
                    enableReload();
                }
            }
        }).start();
    }

    private void enableReload() {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                refreshButton.setEnabled(true);
            }
        });
    }

    private void displayErrorMessage(final String message) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                errView.setVisibility(View.VISIBLE);
                errView.setText(message);
            }
        });
    }

    private void displayList(final MessageEvent evt) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                try {
                    LogEntries log = new LogEntries(evt.getData());
                    while (listView.getChildCount() > 1) {
                        listView.removeViewAt(1);
                    }
                    errView.setVisibility(View.GONE);
                    addLogViews(log.getEntries());
                } catch (JSONException e) {
                    errView.setVisibility(View.VISIBLE);
                    errView.setText("internal JSON error: " + e.toString());
                }
            }
        });
    }

    private void addLogViews(LogEntries.Entry[] entries) {
        for (LogEntries.Entry e : entries) {
            listView.addView(new LogEntry(this, e.service, e.message));
        }
    }
}
