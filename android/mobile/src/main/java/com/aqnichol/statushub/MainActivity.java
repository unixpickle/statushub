package com.aqnichol.statushub;

import android.app.AlertDialog;
import android.content.DialogInterface;
import android.content.SharedPreferences;
import android.support.v7.app.AppCompatActivity;
import android.os.Bundle;
import android.util.Log;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;

import com.google.android.gms.common.ConnectionResult;
import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.PutDataMapRequest;
import com.google.android.gms.wearable.PutDataRequest;
import com.google.android.gms.wearable.Wearable;

public class MainActivity extends AppCompatActivity {

    private final static String TAG = "StatusHub";

    private EditText rootURL;
    private EditText password;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        final Button b = (Button)findViewById(R.id.saveButton);
        b.setOnClickListener(new View.OnClickListener() {
            public void onClick(View v) {
                saveSettings();
            }
        });
        password = (EditText)findViewById(R.id.password);
        rootURL = (EditText)findViewById(R.id.rootURL);
        loadSettings();
    }

    private void loadSettings() {
        SharedPreferences prefs = getSharedPreferences("shhost", 0);
        rootURL.setText(prefs.getString("rootURL", ""));
        password.setText(prefs.getString("password", ""));
    }

    private void saveSettings() {
        SharedPreferences prefs = getSharedPreferences("shhost", 0);
        SharedPreferences.Editor e = prefs.edit();
        e.putString("rootURL", rootURL.getText().toString());
        e.putString("password", password.getText().toString());
        e.commit();
    }
}
