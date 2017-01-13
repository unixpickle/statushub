package com.aqnichol.statushub;

import com.aqnichol.watchcomm.Sender;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.WearableListenerService;

public class RefreshService extends WearableListenerService {
    @Override
    public void onMessageReceived(MessageEvent messageEvent) {
        Sender c = new Sender(getApplicationContext());
        try {
            c.connect();
            c.sendMessage(messageEvent.getSourceNodeId(), "/listing", null);
        } catch (Sender.CommException e) {
        } finally {
            c.disconnect();
        }
    }
}
